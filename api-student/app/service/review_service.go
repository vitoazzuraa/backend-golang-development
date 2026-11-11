package service

import (
	"errors"
	"strconv"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
)

type ReviewService struct {
	reviews repository.ReviewRepository
}

func NewReviewService(reviews repository.ReviewRepository) *ReviewService {
	return &ReviewService{reviews: reviews}
}

func (s *ReviewService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courseID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.CreateReviewRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	review := model.Review{
		CourseID: courseID,
		UserID:   current.UserID,
		Username: current.Username,
		Rating:   req.Rating,
		Comment:  req.Comment,
		Tags:     req.Tags,
	}

	created, err := s.reviews.Create(ctx, review)

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Created(c, "ulasan berhasil dibuat", created,
		"/api/v1/reviews/"+created.ID)
}

func (s *ReviewService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	review, err := s.reviews.FindByID(ctx, c.Params("id"))

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Ok(c, "ulasan ditemukan", review)
}

func (s *ReviewService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	courseID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	query := model.ReviewQuery{
		CourseID: courseID,
		Tag:      c.Query("tag"),
		Before:   c.Query("before"),
		Limit:    c.QueryInt("limit", 10),
	}

	reviews, err := s.reviews.FindByCourse(ctx, query)

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Ok(c, "daftar ulasan berhasil diambil", reviews)
}

func (s *ReviewService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.UpdateReviewRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if req.Rating == nil && req.Comment == nil && req.Tags == nil {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	review, err := s.reviews.UpdateOwn(ctx, c.Params("id"), current.UserID, req)

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Ok(c, "ulasan berhasil diubah", review)
}

func (s *ReviewService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	anyOwner := current.Role == "admin"

	if err := s.reviews.Delete(ctx, c.Params("id"), current.UserID, anyOwner); err != nil {
		return translateReviewError(err)
	}

	return helper.NoContent(c)
}

func (s *ReviewService) AddReply(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.ReplyRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	reply := model.Reply{
		UserID:   current.UserID,
		Username: current.Username,
		Body:     req.Body,
	}

	review, err := s.reviews.AddReply(ctx, c.Params("id"), reply)

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Created(c, "balasan berhasil ditambahkan", review,
		"/api/v1/reviews/"+c.Params("id"))
}

func (s *ReviewService) MarkHelpful(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	review, err := s.reviews.MarkHelpful(ctx, c.Params("id"), current.UserID)

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Ok(c, "ulasan ditandai membantu", review)
}

func (s *ReviewService) Summary(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	courseID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	summary, err := s.reviews.Summary(ctx, courseID)

	if err != nil {
		return translateReviewError(err)
	}

	return helper.Ok(c, "ringkasan ulasan berhasil diambil", summary)
}

func translateReviewError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("ulasan tidak ditemukan")
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("format id tidak valid")
	case errors.Is(err, repository.ErrNotOwner):
		return helper.Forbidden("Anda bukan pemilik ulasan ini")
	case errors.Is(err, repository.ErrAlreadyVoted):
		return helper.Conflict("Anda sudah pernah menandai ulasan ini")
	case errors.Is(err, repository.ErrReplyLimit):
		return helper.NewError(fiber.StatusConflict, "REPLY_LIMIT",
			"jumlah balasan sudah mencapai batas "+strconv.Itoa(repository.MaxRepliesPerReview))
	default:
		return helper.Internal(err)
	}
}
