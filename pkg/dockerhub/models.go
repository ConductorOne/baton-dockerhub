package dockerhub

type BaseResource struct {
	Id string `json:"id"`
}

type PaginationData struct {
	Count int    `json:"count"`
	Next  string `json:"next"`
}

type User struct {
	BaseResource
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Team struct {
	BaseResource
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TeamMember struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
