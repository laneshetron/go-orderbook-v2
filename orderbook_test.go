// Copyright 2024 Lane A. Shetron
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package orderbook

import (
	"fmt"
	"math/rand"
	"testing"
)

// ~1.2us for 1M records
func BenchmarkBidBookInsertWorstCase(b *testing.B) {
	ob := NewOrderBook()
	for n := 0; n < b.N; n++ {
		ob.Insert(n, BID, 1.0+float32(n), 1)
	}
}

// ~0.5us for 2M records
func BenchmarkBidBookInsertAverage(b *testing.B) {
	ob := NewOrderBook()
	for n := 0; n < b.N; n++ {
		ob.Insert(n, BID, float32(int(rand.Float32()*10000))/10000, 1)
	}
}

func TestAskBook(t *testing.T) {
	orders := []struct {
		Id    int
		Price float32
		Peek  float32
	}{
		{1, 123.45, 123.45},
		{2, 155.45, 123.45},
		{3, 122.00, 122.00},
		{8, 122.00, 122.00},
		{9, 122.00, 122.00},
		{4, 136.00, 122.00},
		{5, 121.00, 121.00},
		{10, 121.00, 121.00},
		{6, 333.00, 121.00},
		{7, 120.999, 120.999},
	}
	ob := NewOrderBook()
	for _, order := range orders {
		t.Run(fmt.Sprintf("%d-%f", order.Id, order.Price), func(t *testing.T) {
			ob.Insert(order.Id, ASK, order.Price, 1)
			if ob.AskBook.Peek().Price != order.Peek {
				t.Errorf("Expected lowest ask %f, got %f", order.Peek, ob.AskBook.Peek().Price)
			}
		})
	}
	expected := []float32{120.999, 121.00, 121.00, 122.00, 122.0, 122.00, 123.45, 136.00, 155.45, 333.00}
	for ob.AskBook.Len() > 0 {
		t.Run(fmt.Sprintf("next-lowest-%f", expected[0]), func(t *testing.T) {
			o := ob.AskBook.Pop().Peek()
			if o.Price != expected[0] {
				t.Errorf("Expected next lowest ask %f, got %f", expected[0], o.Price)
			}
			expected = expected[1:]
		})
	}
}

func TestBidBook(t *testing.T) {
	orders := []struct {
		Id    int
		Price float32
		Peek  float32
	}{
		{1, 123.45, 123.45},
		{2, 155.45, 155.45},
		{3, 122.00, 155.45},
		{4, 136.00, 155.45},
		{5, 121.00, 155.45},
		{6, 333.00, 333.00},
		{7, 120.999, 333.00},
	}
	ob := NewOrderBook()
	for _, order := range orders {
		t.Run(fmt.Sprintf("%d-%f", order.Id, order.Price), func(t *testing.T) {
			ob.Insert(order.Id, BID, order.Price, 1)
			if ob.BidBook.Peek().Price != order.Peek {
				t.Errorf("Expected highest bid %f, got %f", order.Peek, ob.BidBook.Peek().Price)
			}
		})
	}
	expected := []float32{333.00, 155.45, 136.00, 123.45, 122.00, 121.00, 120.999}
	for ob.BidBook.Len() > 0 {
		t.Run(fmt.Sprintf("next-highest-%f", expected[0]), func(t *testing.T) {
			o := ob.BidBook.Pop().Peek()
			if o.Price != expected[0] {
				t.Errorf("Expected next highest bid %f, got %f", expected[0], o.Price)
			}
			expected = expected[1:]
		})
	}
}
