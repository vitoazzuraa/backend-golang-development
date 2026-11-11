package repository

import (
	"context"
	"errors"
	"fmt"

	"backend-go/api-student/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, portfolio model.Portfolio) error
	AddProject(ctx context.Context, studentID int, project model.Project) error
	RemoveProject(ctx context.Context, studentID int, projectID string) error
	TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyStat, error)
}

var (
	ErrAlreadyExists   = errors.New("portofolio sudah ada")
	ErrProjectLimit    = errors.New("jumlah proyek sudah mencapai batas")
	ErrProjectMissing  = errors.New("proyek tidak ditemukan")
	ErrInvalidObjectID = errors.New("format id proyek tidak valid")
)

type portfolioDocument struct {
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
}

type projectDocument struct {
	ID           string   `bson:"id"`
	Title        string   `bson:"title"`
	Technologies []string `bson:"technologies"`
	Link         string   `bson:"link"`
	Year         int      `bson:"year"`
}

type certificateDocument struct {
	Name     string `bson:"name"`
	Issuer   string `bson:"issuer"`
	IssuedAt string `bson:"issued_at"`
}

type portfolioMongoRepository struct {
	collection *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{collection: db.Collection("student_portfolios")}
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("student_portfolios_student_unique"),
	})

	if err != nil {
		return fmt.Errorf("menyiapkan index portofolio: %w", err)
	}

	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "projects.technologies", Value: 1}},
	})

	if err != nil {
		return fmt.Errorf("menyiapkan index teknologi: %w", err)
	}

	return nil
}

func toDocument(p model.Portfolio) portfolioDocument {
	doc := portfolioDocument{
		StudentID: p.StudentID,
		Headline:  p.Headline,
		Bio:       p.Bio,
		Projects:  []projectDocument{},
	}

	for _, project := range p.Projects {
		doc.Projects = append(doc.Projects, projectDocument{
			ID:           project.ID,
			Title:        project.Title,
			Technologies: project.Technologies,
			Link:         project.Link,
			Year:         project.Year,
		})
	}

	doc.Certificates = []certificateDocument{}

	for _, certificate := range p.Certificates {
		doc.Certificates = append(doc.Certificates, certificateDocument{
			Name:     certificate.Name,
			Issuer:   certificate.Issuer,
			IssuedAt: certificate.IssuedAt,
		})
	}

	return doc
}

func toModel(doc portfolioDocument) model.Portfolio {
	result := model.Portfolio{
		StudentID:    doc.StudentID,
		Headline:     doc.Headline,
		Bio:          doc.Bio,
		Projects:     []model.Project{},
		Certificates: []model.Certificate{},
	}

	for _, project := range doc.Projects {
		result.Projects = append(result.Projects, model.Project{
			ID:           project.ID,
			Title:        project.Title,
			Technologies: project.Technologies,
			Link:         project.Link,
			Year:         project.Year,
		})
	}

	for _, certificate := range doc.Certificates {
		result.Certificates = append(result.Certificates, model.Certificate{
			Name:     certificate.Name,
			Issuer:   certificate.Issuer,
			IssuedAt: certificate.IssuedAt,
		})
	}

	return result
}

func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument

	err := r.collection.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}

		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}

	return toModel(doc), nil
}

func (r *portfolioMongoRepository) Upsert(ctx context.Context, portfolio model.Portfolio) error {
	doc := toDocument(portfolio)

	_, err := r.collection.ReplaceOne(ctx,
		bson.M{"student_id": portfolio.StudentID},
		doc,
		options.Replace().SetUpsert(true))

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAlreadyExists
		}

		return fmt.Errorf("menyimpan portofolio: %w", err)
	}

	return nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, project model.Project) error {
	doc := projectDocument{
		ID:           project.ID,
		Title:        project.Title,
		Technologies: project.Technologies,
		Link:         project.Link,
		Year:         project.Year,
	}

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"student_id": studentID, "projects.19": bson.M{"$exists": false}},
		bson.M{"$push": bson.M{"projects": doc}})

	if err != nil {
		return fmt.Errorf("menambah proyek: %w", err)
	}

	if result.MatchedCount > 0 {
		return nil
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"student_id": studentID})

	if err != nil {
		return fmt.Errorf("menghitung portofolio: %w", err)
	}

	if count == 0 {
		return ErrNotFound
	}

	return ErrProjectLimit
}

func (r *portfolioMongoRepository) RemoveProject(ctx context.Context, studentID int, projectID string) error {
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"student_id": studentID},
		bson.M{"$pull": bson.M{"projects": bson.M{"id": projectID}}})

	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	if result.ModifiedCount == 0 {
		return ErrProjectMissing
	}

	return nil
}

func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyStat, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$projects"}},
		bson.D{{Key: "$unwind", Value: "$projects.technologies"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$projects.technologies"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}, {Key: "_id", Value: 1}}}},
		bson.D{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)

	if err != nil {
		return nil, fmt.Errorf("menghitung teknologi: %w", err)
	}

	defer cursor.Close(ctx)

	result := []model.TechnologyStat{}

	for cursor.Next(ctx) {
		var row struct {
			ID    string `bson:"_id"`
			Total int    `bson:"total"`
		}

		if err := cursor.Decode(&row); err != nil {
			return nil, fmt.Errorf("membaca hasil teknologi: %w", err)
		}

		result = append(result, model.TechnologyStat{
			Technology: row.ID,
			Total:      row.Total,
		})
	}

	return result, cursor.Err()
}
