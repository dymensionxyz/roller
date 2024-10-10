package migrations

import "time"

// Commit struct is used to retrieve a timestamp of a git commit and
// by comparing the timestamp of the installed version with the timestamp
// of the previous version, determine whether to perform migration
type Commit struct {
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

// Tag struct is used to retrieve a timestamp of a git tag and
// by comparing the timestamp of the installed version with the timestamp
// of the previous version, determine whether to perform migration
type Tag struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

// Release struct is used to retrieve a timestamp of a git release and
// by comparing the timestamp of the installed version with the timestamp
// of the previous version, determine whether to perform migration
type Release struct {
	TagName string `json:"tag_name"`
}
