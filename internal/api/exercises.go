package api

import "fmt"

func (c *Client) GetExercises(trackSlug string) ([]Exercise, error) {
	url := fmt.Sprintf("%s/tracks/%s/exercises", websiteAPI, trackSlug)
	cacheKey := fmt.Sprintf("exercises:%s", trackSlug)
	data, err := c.getCached(cacheKey, url, true)
	if err != nil {
		return nil, err
	}

	resp, err := decode[ExercisesResponse](data)
	if err != nil {
		return nil, err
	}

	return resp.Exercises, nil
}
