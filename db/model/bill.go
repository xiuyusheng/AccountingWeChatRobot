package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Bill struct {
	Time        Time     `json:"time" gorm:"column:Time;primaryKey"`
	Consumption float64  `json:"consumption" gorm:"column:Consumption"`
	Note        string   `json:"note" gorm:"column:Note"`
	Consumer    Consumer `json:"consumer" gorm:"column:Consumer"`
}

type Time struct {
	time.Time
}

func (b *Time) Scan(v interface{}) error {
	TInt, ok := v.(int64)
	if !ok {
		return errors.New("Time is not a number")
	}
	b.Time = time.Unix(TInt, 0)
	return nil
}

func (b Time) Value() (driver.Value, error) {
	return b.Unix(), nil
}

type Consumer []string

func (c *Consumer) Scan(v interface{}) error {
	data, ok := v.([]uint8)
	if !ok {
		return errors.New("Consumer is not a string")
	}
	return json.Unmarshal([]byte(data), c)
}

func (c Consumer) Value() (driver.Value, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return data, nil
}
