package main

import (
	"fmt"

	"github.com/binadel/esdigo/json"
)

func main() {
	//p := &Product{}
	//p.Title.Set("MacBook Air M5")
	//p.Price.SetNull()
	//p.IsPublished.Set(true)
	//
	//r := &ProductResponse{}
	//r.Product.Set(p)
	//
	//data, err := r.MarshalJSON()
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Println(string(data))
	//}
	//
	//r1 := &ProductResponse{}
	//err = r1.UnmarshalJSON(data)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(r1.Product.Value)

	js := `{"title": "ps5", "isPublished": false, "email": "s user@email.com"}`
	p := &Product{}
	err := p.UnmarshalJSON([]byte(js))
	if err != nil {
		fmt.Println(err)
		return
	}

	validator := NewProductValidator()
	validated := validator.Validate(p)

	fmt.Println(validated.Title.Value)
	fmt.Println(validated.IsPublished.Value)
	fmt.Println(validated.Email.Value)

	w := json.NewWriter(128)
	validated.Title.WriteJSON(w)
	validated.IsPublished.WriteJSON(w)
	validated.Email.WriteJSON(w)
	fmt.Println(string(w.Bytes()))
}
