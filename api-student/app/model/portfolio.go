package model

type Portfolio struct {
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
}

type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link"`
	Year         int      `json:"year"`
}

type Certificate struct {
	Name     string `json:"name"`
	Issuer   string `json:"issuer"`
	IssuedAt string `json:"issued_at"`
}

type SavePortfolioRequest struct {
	Headline     string        `json:"headline" validate:"required,min=5,max=120"`
	Bio          string        `json:"bio" validate:"required,min=10,max=1000"`
	Projects     []Project     `json:"projects" validate:"omitempty,max=20,dive"`
	Certificates []Certificate `json:"certificates" validate:"omitempty,max=30,dive"`
}

type AddProjectRequest struct {
	Title        string   `json:"title" validate:"required,min=3,max=120"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=1,max=30"`
	Link         string   `json:"link" validate:"omitempty,max=300"`
	Year         int      `json:"year" validate:"required,min=1900,max=2100"`
}

type TechnologyStat struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
