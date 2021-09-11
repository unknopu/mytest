package models

import (
	// "time"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// "github.com/dgrijalva/jwt-go"
)

type User struct {
	// ID     primitive.ObjectID `json:"_id,omitempty"`
	UserName  string `json: "username"`
	LastMatch History

	// 0 - free, 1 - being challenged, 2 - challenging
	Status    int `json: "status"`
	W        int  `json: "win"`
	L        int  `json: "lose"`
	N		int	  `json: neutral`
}

func (u *User) MakeDefaultVal() {
	u.Status = 0
	u.W = 0
	u.L = 0
}
// func (u *User) GetToken() (string, error){
// 	claims := jwt.MapClaims{}
// 	claims["authorized"] = true
// 	claims["user_id"] = u.UserName
// 	claims["exp"] = time.Now().Add(time.Hour*24).Unix()
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
// 	Token, err := token.SignedString("tokenSecret")
// 	return Token, err
// }

// 0-tie, 1-win, 2-lose
// 0-hammer, 1-paper, 2-scissor
type History struct {
	UserName  string `json: "username"`
	Win	int	`json: "iswin"`
}
func (h *History) AddData(name string, result int) {
	h.UserName = name
	h.Win = result
}

type Record struct{
	SendBy string `json: "sendby"`
	SendTo string `json: "sendto"`
	Choise int `json: "choise"`
}
// func (r *Record) AddInfo(s string, recv string, choice int) {
// 	r.SendBy = s
// 	r.SendTo = recv
// 	r.Choise = choice
// }


