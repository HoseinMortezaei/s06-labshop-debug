// Command labshop prints the LabShop daily report.
//
//	go run . ../data/orders.csv
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"reflect"

	"labshop/pricing"
)

const header = "order  customer     subtotal  discount       tax  shipping       total"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <orders.csv>")
		os.Exit(2)
	}
	orders, err := readOrders(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("LabShop daily report")
	fmt.Println(header)
	invalid := 0
	for _, o := range orders {
		//fmt.Println(reflect.TypeOf(o))
		//fmt.Println("%+d",o)
		b := pricing.PriceOrder(o)
		fmt.Printf("%5d  %-10s %10d %9d %9d %9d %11d\n",
			o.ID, o.Customer, b.Subtotal, b.Discount, b.Tax, b.Shipping, b.Total)
		if o.CouponCode != "" && pricing.FindCoupon(o.CouponCode) == nil {
			invalid++
		}
	}
	fmt.Printf("orders: %d\n", len(orders))
	fmt.Printf("invalid coupons: %d\n", invalid)
	fmt.Printf("grand total: %d\n", grandTotal(orders))
}

func grandTotal(orders []pricing.Order) int64 {
	var sum int64
	for _, o := range orders {
		sum += pricing.PriceOrder(o).Total
	}
	return sum
}

func readOrders(path string) ([]pricing.Order, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var orders []pricing.Order
	sc := bufio.NewScanner(f)
	sc.Scan() // header
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) != 4 {
			return nil, fmt.Errorf("bad line: %q", line)
		}
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, err
		}
		o := pricing.Order{ID: id, Customer: fields[1], CouponCode: fields[2]}
		for _, part := range strings.Split(fields[3], "|") {
			p := strings.Split(part, ":")
			price, err := strconv.ParseInt(p[1], 10, 64)
			if err != nil {
				return nil, err
			}
			qty, err := strconv.ParseInt(p[2], 10, 64)
			if err != nil {
				return nil, err
			}
			o.Items = append(o.Items, pricing.Item{SKU: p[0], UnitPrice: price, Qty: qty})
		}
		orders = append(orders, o)
	}
	return orders, sc.Err()
}
