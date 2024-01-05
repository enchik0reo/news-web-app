package models

import (
	"time"
)

type Article struct {
	UserName   string    `json:"user_name"`
	SourceName string    `json:"source_name"`
	Title      string    `json:"title"`
	Link       string    `json:"link"`
	Excerpt    string    `json:"excerpt"`
	ImageURL   string    `json:"image_url"`
	PostedAt   time.Time `json:"posted_at"`
}

type Art struct {
	Link    string
	Content string
}

/*
type User struct {
	ID   int64
	Name string
} */

/* func (a Article) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	err := enc.Encode(a.UserName)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(a.SourceName)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(a.Title)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(a.Link)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(a.Excerpt)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(a.ImageURL)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(a.PostedAt)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (a *Article) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&a.UserName)
	if err != nil {
		return err
	}
	err = dec.Decode(&a.SourceName)
	if err != nil {
		return err
	}
	err = dec.Decode(&a.Title)
	if err != nil {
		return err
	}
	err = dec.Decode(&a.Link)
	if err != nil {
		return err
	}
	err = dec.Decode(&a.Excerpt)
	if err != nil {
		return err
	}
	err = dec.Decode(&a.ImageURL)
	if err != nil {
		return err
	}
	err = dec.Decode(&a.PostedAt)
	if err != nil {
		return err
	}

	return nil
}
*/
