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
	code = strings.ToUpper(strings.TrimSpace(code))
	return coupons[code]
}

func PriceOrder(o Order) Bill {
	var subtotal int64
	for _, it := range o.Items {
		subtotal += it.UnitPrice * it.Qty
	}
	bulk := bulkDiscount(o.Items)
	coupon := couponDiscount(o, subtotal-bulk)
	taxable := subtotal - bulk - coupon
	tax := TaxFor(taxable)
	shipping := ShippingFor(taxable)
	return Bill{subtotal, bulk + coupon, tax, shipping, taxable + tax + shipping}
}

func sumQtyBySKU(items []Item) map[string]Item {
	sums := make(map[string]Item, len(items))
	for _, it := range items {
		// should return an error if prices don't match but we ignore
		if existing, ok := sums[it.SKU]; ok {
			existing.Qty += it.Qty
			sums[it.SKU] = existing
		} else {
			sums[it.SKU] = it
		}
	}
	return sums
}

// bulkDiscount gives 10% off every line that has BulkMinQty units or more.
func bulkDiscount(items []Item) int64 {
	var discount int64
	sum := sumQtyBySKU(items)
	for _, it := range sum {
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
	c := FindCoupon(o.CouponCode)
	if c == nil {
		return 0
	}
	if c.Kind == "fixed" {
		return min(c.Value, amount)
	}
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
