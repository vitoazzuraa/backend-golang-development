package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"backend-go/api-student/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const MaxRepliesPerReview = 50

var (
	ErrInvalidID    = errors.New("format id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak boleh menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah pernah menilai ulasan ini")
	ErrReplyLimit   = errors.New("jumlah balasan sudah mencapai batas")
)

type ReviewRepository interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, r model.Review) (model.Review, error)
	FindByID(ctx context.Context, id string) (model.Review, error)
	FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error)
	UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error)
	Delete(ctx context.Context, id string, userID int, anyOwner bool) error
	AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error)
	MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error)
	Summary(ctx context.Context, courseID int) (model.ReviewSummary, error)
}

type reviewDocument struct {
	ID           bson.ObjectID   `bson:"_id,omitempty"`
	CourseID     int             `bson:"course_id"`
	UserID       int             `bson:"user_id"`
	Username     string          `bson:"username"`
	Rating       int             `bson:"rating"`
	Comment      string          `bson:"comment"`
	Tags         []string        `bson:"tags"`
	Replies      []replyDocument `bson:"replies"`
	HelpfulCount int             `bson:"helpful_count"`
	HelpfulBy    []int           `bson:"helpful_by"`
	CreatedAt    time.Time       `bson:"created_at"`
	UpdatedAt    time.Time       `bson:"updated_at"`
}

type replyDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	UserID    int           `bson:"user_id"`
	Username  string        `bson:"username"`
	Body      string        `bson:"body"`
	CreatedAt time.Time     `bson:"created_at"`
}

func (d reviewDocument) toModel() model.Review {
	replies := make([]model.Reply, 0, len(d.Replies))

	for _, r := range d.Replies {
		replies = append(replies, model.Reply{
			ID: r.ID.Hex(), UserID: r.UserID, Username: r.Username,
			Body: r.Body, CreatedAt: r.CreatedAt,
		})
	}

	tags := d.Tags

	if tags == nil {
		tags = []string{}
	}

	return model.Review{
		ID: d.ID.Hex(), CourseID: d.CourseID, UserID: d.UserID, Username: d.Username,
		Rating: d.Rating, Comment: d.Comment, Tags: tags, Replies: replies,
		HelpfulCount: d.HelpfulCount, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type reviewMongoRepository struct {
	coll *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewMongoRepository{coll: db.Collection("course_reviews")}
}

func now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func parseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)

	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}

	return oid, nil
}

func (r *reviewMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "course_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("course_user_unique")},
		{Keys: bson.D{{Key: "course_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "tags", Value: 1}}},
	})

	if err != nil {
		return fmt.Errorf("menyiapkan index ulasan: %w", err)
	}

	return nil
}

func (r *reviewMongoRepository) Create(ctx context.Context, input model.Review) (model.Review, error) {
	doc := reviewDocument{
		ID:        bson.NewObjectID(),
		CourseID:  input.CourseID,
		UserID:    input.UserID,
		Username:  input.Username,
		Rating:    input.Rating,
		Comment:   input.Comment,
		Tags:      input.Tags,
		Replies:   []replyDocument{},
		HelpfulBy: []int{},
		CreatedAt: now(),
		UpdatedAt: now(),
	}

	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.Review{}, ErrDuplicate
		}

		return model.Review{}, fmt.Errorf("menyimpan ulasan: %w", err)
	}

	return doc.toModel(), nil
}

func (r *reviewMongoRepository) FindByID(ctx context.Context, id string) (model.Review, error) {
	oid, err := parseID(id)

	if err != nil {
		return model.Review{}, err
	}

	var doc reviewDocument

	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Review{}, ErrNotFound
		}

		return model.Review{}, fmt.Errorf("mengambil ulasan: %w", err)
	}

	return doc.toModel(), nil
}

func (r *reviewMongoRepository) FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error) {
	filter := bson.M{"course_id": q.CourseID}

	if q.Tag != "" {
		filter["tags"] = q.Tag
	}

	if q.Before != "" {
		oid, err := parseID(q.Before)

		if err != nil {
			return nil, err
		}

		filter["_id"] = bson.M{"$lt": oid}
	}

	limit := q.Limit

	if limit < 1 || limit > 50 {
		limit = 10
	}

	cursor, err := r.coll.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(int64(limit)))

	if err != nil {
		return nil, fmt.Errorf("mengambil daftar ulasan: %w", err)
	}

	defer cursor.Close(ctx)

	result := []model.Review{}

	for cursor.Next(ctx) {
		var doc reviewDocument

		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("membaca ulasan: %w", err)
		}

		result = append(result, doc.toModel())
	}

	return result, cursor.Err()
}

func (r *reviewMongoRepository) UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error) {
	oid, err := parseID(id)

	if err != nil {
		return model.Review{}, err
	}

	set := bson.M{"updated_at": now()}

	if req.Rating != nil {
		set["rating"] = *req.Rating
	}

	if req.Comment != nil {
		set["comment"] = *req.Comment
	}

	if req.Tags != nil {
		set["tags"] = *req.Tags
	}

	result, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": oid, "user_id": userID},
		bson.M{"$set": set})

	if err != nil {
		return model.Review{}, fmt.Errorf("mengubah ulasan: %w", err)
	}

	if result.MatchedCount == 0 {
		if _, findErr := r.FindByID(ctx, id); findErr != nil {
			return model.Review{}, findErr
		}

		return model.Review{}, ErrNotOwner
	}

	return r.FindByID(ctx, id)
}

func (r *reviewMongoRepository) Delete(ctx context.Context, id string, userID int, anyOwner bool) error {
	oid, err := parseID(id)

	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid}

	if !anyOwner {
		filter["user_id"] = userID
	}

	result, err := r.coll.DeleteOne(ctx, filter)

	if err != nil {
		return fmt.Errorf("menghapus ulasan: %w", err)
	}

	if result.DeletedCount == 0 {
		if _, findErr := r.FindByID(ctx, id); findErr != nil {
			return findErr
		}

		return ErrNotOwner
	}

	return nil
}

func (r *reviewMongoRepository) AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error) {
	oid, err := parseID(id)

	if err != nil {
		return model.Review{}, err
	}

	doc := replyDocument{
		ID:        bson.NewObjectID(),
		UserID:    reply.UserID,
		Username:  reply.Username,
		Body:      reply.Body,
		CreatedAt: now(),
	}

	result, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": oid, fmt.Sprintf("replies.%d", MaxRepliesPerReview-1): bson.M{"$exists": false}},
		bson.M{"$push": bson.M{"replies": doc}, "$set": bson.M{"updated_at": now()}})

	if err != nil {
		return model.Review{}, fmt.Errorf("menambah balasan: %w", err)
	}

	if result.MatchedCount == 0 {
		if _, findErr := r.FindByID(ctx, id); findErr != nil {
			return model.Review{}, findErr
		}

		return model.Review{}, ErrReplyLimit
	}

	return r.FindByID(ctx, id)
}

func (r *reviewMongoRepository) MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error) {
	oid, err := parseID(id)

	if err != nil {
		return model.Review{}, err
	}

	result, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": oid, "user_id": bson.M{"$ne": userID}, "helpful_by": bson.M{"$ne": userID}},
		bson.M{"$addToSet": bson.M{"helpful_by": userID}, "$inc": bson.M{"helpful_count": 1}})

	if err != nil {
		return model.Review{}, fmt.Errorf("menandai ulasan membantu: %w", err)
	}

	if result.MatchedCount == 0 {
		existing, findErr := r.FindByID(ctx, id)

		if findErr != nil {
			return model.Review{}, findErr
		}

		if existing.UserID == userID {
			return model.Review{}, ErrOwnReview
		}

		return model.Review{}, ErrAlreadyVoted
	}

	return r.FindByID(ctx, id)
}

func (r *reviewMongoRepository) Summary(ctx context.Context, courseID int) (model.ReviewSummary, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"course_id": courseID}}},
		bson.D{{Key: "$facet", Value: bson.M{
			"Value": mongo.Pipeline{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "Total", Value: bson.D{{Key: "$sum", Value: 1}}},
					{Key: "rata", Value: bson.D{{Key: "$avg", Value: "$rating"}}},
				}}},
			},
			"Spread": mongo.Pipeline{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$rating"},
					{Key: "Total", Value: bson.D{{Key: "$sum", Value: 1}}},
				}}},
			},
			"tag": mongo.Pipeline{
				bson.D{{Key: "$unwind", Value: "$tags"}},
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$tags"},
					{Key: "Total", Value: bson.D{{Key: "$sum", Value: 1}}},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{{Key: "Total", Value: -1}, {Key: "_id", Value: 1}}}},
				bson.D{{Key: "$limit", Value: 5}},
			},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)

	if err != nil {
		return model.ReviewSummary{}, fmt.Errorf("menghitung ringkasan: %w", err)
	}

	defer cursor.Close(ctx)

	var hasil []struct {
		Value []struct {
			Total int     `bson:"Total"`
			Rata  float64 `bson:"rata"`
		} `bson:"Value"`
		Spread []struct {
			ID    int `bson:"_id"`
			Total int `bson:"Total"`
		} `bson:"Spread"`
		Tag []struct {
			ID    string `bson:"_id"`
			Total int    `bson:"Total"`
		} `bson:"tag"`
	}

	if err := cursor.All(ctx, &hasil); err != nil {
		return model.ReviewSummary{}, fmt.Errorf("membaca ringkasan: %w", err)
	}

	summary := model.ReviewSummary{
		CourseID:     courseID,
		Distribution: map[string]int{},
		TopTags:      []model.TagCount{},
	}

	if len(hasil) == 0 {
		return summary, nil
	}

	for _, item := range hasil[0].Value {
		summary.Count = item.Total
		summary.AverageRating = math.Round(item.Rata*100) / 100
	}

	for _, item := range hasil[0].Spread {
		summary.Distribution[fmt.Sprintf("%d", item.ID)] = item.Total
	}

	for _, item := range hasil[0].Tag {
		summary.TopTags = append(summary.TopTags, model.TagCount{Tag: item.ID, Total: item.Total})
	}

	for bintang := 1; bintang <= 5; bintang++ {
		kunci := fmt.Sprintf("%d", bintang)

		if _, ada := summary.Distribution[kunci]; !ada {
			summary.Distribution[kunci] = 0
		}
	}

	return summary, nil
}
