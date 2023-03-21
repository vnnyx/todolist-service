package web

type ActivityDTO struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ActivityCreateRequest struct {
	Title string `json:"title"`
	Email string `json:"email"`
}

type ActivityUpdateRequest struct {
	ID    int64
	Title string `json:"title"`
}
