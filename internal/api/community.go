package api

import "fmt"

// GetCommunitySolutions fetches published community solutions for an exercise.
func (c *Client) GetCommunitySolutions(trackSlug, exerciseSlug string) ([]CommunitySolution, error) {
	url := fmt.Sprintf("%s/tracks/%s/exercises/%s/community_solutions", websiteAPI, trackSlug, exerciseSlug)
	data, err := c.get(url, true)
	if err != nil {
		return nil, err
	}

	resp, err := decode[CommunitySolutionsResponse](data)
	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// GetCommunitySolutionFiles fetches the files for a specific community solution.
func (c *Client) GetCommunitySolutionFiles(trackSlug, exerciseSlug, handle string) ([]CommunitySolutionFile, error) {
	url := fmt.Sprintf("%s/tracks/%s/exercises/%s/community_solutions/%s", websiteAPI, trackSlug, exerciseSlug, handle)
	data, err := c.get(url, true)
	if err != nil {
		return nil, err
	}

	// The single solution endpoint wraps files differently
	var resp struct {
		Solution struct {
			Files []CommunitySolutionFile `json:"files"`
		} `json:"solution"`
	}
	if err := decodeInto(data, &resp); err != nil {
		return nil, err
	}

	return resp.Solution.Files, nil
}
