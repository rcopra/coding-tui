package api

func (c *Client) GetTracks() ([]Track, error) {
	url := websiteAPI + "/tracks"
	data, err := c.getCached("tracks", url, true)
	if err != nil {
		return nil, err
	}

	resp, err := decode[TracksResponse](data)
	if err != nil {
		return nil, err
	}

	return resp.Tracks, nil
}
