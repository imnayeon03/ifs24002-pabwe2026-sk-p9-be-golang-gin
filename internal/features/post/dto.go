package post

var Categories = []string{"Teknologi", "Pendidikan", "Gaya Hidup", "Bisnis", "Opini", "Umum"}

type PostRequest struct {
	Title          string `json:"title" binding:"required"`
	Category       string `json:"category"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content"`
	CoverImagePath string `json:"cover_image_path"`
}

type Stats struct {
	Total        int64 `json:"total"`
	TotalTerbit  int64 `json:"total_terbit"`
	BelumSelesai int64 `json:"belum_selesai"`
}

type Filter struct {
	Query    string
	Category string
}
