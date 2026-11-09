// Package pricing holds the LabShop pricing rules. All amounts are in toman.
package pricing

import (
	"strings"
)

const (
	BulkMinQty       = 10
	BulkPercent      = 10
	TaxPercent       = 9
	ShippingFee      = 40_000
	FreeShippingFrom = 1_000_000
)

type Item struct {
	SKU       string
	UnitPrice int64
	Qty       int64
}

type Order struct {
	ID         int
	Customer   string
	CouponCode string
	Items      []Item
}

type Bill struct {
	Subtotal int64
	Discount int64
	Tax      int64
	Shipping int64
	Total    int64
}

type Coupon struct {
	Kind  string // "percent" or "fixed"
	Value int64
}

var coupons = map[string]*Coupon{
	"OFF10": {Kind: "percent", Value: 10},
	"OFF20": {Kind: "percent", Value: 20},
	"FIX50": {Kind: "fixed", Value: 50_000},
}

func FindCoupon(code string) *Coupon {
	code = strings.TrimSpace(strings.ToUpper(code))
	//fmt.Println(code)
	return coupons[code]
}

func PriceOrder(o Order) Bill {
	var subtotal int64
	for _, it := range o.Items {
		subtotal += it.UnitPrice * it.Qty
	}
	//fmt.Println(1)
	bulk := bulkDiscount(o.Items)
	//fmt.Println(2,bulk)
	coupon := couponDiscount(o, subtotal-bulk)
	//fmt.Println(3)
	taxable := subtotal - bulk - coupon
	tax := TaxFor(taxable)
	//fmt.Println(4)
	shipping := ShippingFor(taxable)
	//fmt.Println(5)
	return Bill{subtotal, bulk + coupon, tax, shipping, taxable + tax + shipping}
}

// bulkDiscount gives 10% off every line that has BulkMinQty units or more.
func bulkDiscount(items []Item) int64 {
	var discount int64
	for _, it := range items {
		if it.Qty >= BulkMinQty {
			discount += it.UnitPrice * it.Qty * BulkPercent / 100
		}
	}
	return discount
}

func couponDiscount(o Order, amount int64) int64 {
	if o.CouponCode == "" {
		return 0
	}
	//fmt.Println("11")
	c := FindCoupon(o.CouponCode)
	//fmt.Println("22",c)
	if(c == nil) {
		return 0
	}
	if c.Kind == "fixed" {
		return min(c.Value, amount)
	}
	//fmt.Println("33")
	return applyPercent(o.Items, c.Value)
}

// applyPercent lowers each unit price so the invoice shows the discounted prices.
func applyPercent(items []Item, percent int64) int64 {
	var discount int64
	for i := range items {
		cut := items[i].UnitPrice * percent / 100
		items[i].UnitPrice -= cut
		discount += cut * items[i].Qty
	}
	return discount
}

// TaxFor returns the VAT, rounded to the nearest toman.
func TaxFor(amount int64) int64 {
	return amount * TaxPercent / 100
}

func ShippingFor(amount int64) int64 {
	if amount >= FreeShippingFrom {
		return 0
	}
	return ShippingFee
}
