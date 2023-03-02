package gexelizer

import (
	"github.com/go-the-way/exl"
)

type Item struct {
	Name  string  `excel:"name"`
	Price float64 `excel:"price"`
}

type Document struct {
	ID          uint    `excel:"id"`
	TotalAmount float64 `excel:"total_amount"`
	Items       []Item  `excel:"items"`
}

func (d Document) Configure(c *exl.WriteConfig) {

}

func main() {
	if err := exl.Write("test.xlsx", []*Document{
		{
			ID:          1,
			TotalAmount: 100,
			Items: []Item{
				{
					Name: "item1",
				},
				{
					Name: "item2",
				},
			},
		},
		{
			ID:          2,
			TotalAmount: 200,
			Items: []Item{
				{
					Name: "item3",
				},
			},
		},
	}); err != nil {
		panic(err)
	}
}
