package api

import "fmt"

func (c *Client) GetExercises(trackSlug string) ([]Exercise, error) {
	url := fmt.Sprintf("%s/tracks/%s/exercises?sideload=solutions", websiteAPI, trackSlug)
	cacheKey := fmt.Sprintf("exercises:%s", trackSlug)
	data, err := c.getCached(cacheKey, url, true)
	if err != nil {
		return nil, err
	}

	resp, err := decode[ExercisesResponse](data)
	if err != nil {
		return nil, err
	}

	// Merge sideloaded solution statuses into exercises
	statusBySlug := make(map[string]string, len(resp.Solutions))
	for _, s := range resp.Solutions {
		statusBySlug[s.Exercise.Slug] = s.Status
	}
	for i := range resp.Exercises {
		resp.Exercises[i].Status = statusBySlug[resp.Exercises[i].Slug]
	}

	return resp.Exercises, nil
}
