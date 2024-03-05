package dockerhub

type BaseResource struct {
	Id string `json:"id"`
}

type PaginationData struct {
	Count int    `json:"count"`
	Next  string `json:"next"`
}

type Organization struct {
	BaseResource
	Name string `json:"orgname"`
}

type User struct {
	BaseResource
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type Team struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Repository struct {
	Name        string `json:"name"`
	NameSpace   string `json:"namespace"`
	Description string `json:"description"`
}

type RepositoryPermission struct {
	TeamId     int    `json:"group_id"`
	TeamName   string `json:"group_name"`
	Permission string `json:"permission"`
}
