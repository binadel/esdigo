package main

import "fmt"

func main() {
	p := &Product{}
	p.Title.Set("MacBook Air M5")
	p.Price.SetNull()
	p.IsPublished.Set(true)

	r := &ProductResponse{}
	r.Product.Set(p)

	data, err := r.MarshalJSON()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(data))
	}

	r1 := &ProductResponse{}
	err = r1.UnmarshalJSON(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r1)
}
