package api

type Track struct {
	Slug                  string   `json:"slug"`
	Title                 string   `json:"title"`
	NumConcepts           int      `json:"num_concepts"`
	NumExercises          int      `json:"num_exercises"`
	NumCompletedExercises int      `json:"num_completed_exercises"`
	NumLearntConcepts     int      `json:"num_learnt_concepts"`
	WebURL                string   `json:"web_url"`
	IconURL               string   `json:"icon_url"`
	Tags                  []string `json:"tags"`
	IsNew                 bool     `json:"is_new"`
	IsJoined              bool     `json:"is_joined"`
	Course                bool     `json:"course"`
	LastTouchedAt         string   `json:"last_touched_at"`
	Links                 Links    `json:"links"`
}

type Links struct {
	Self      string `json:"self"`
	Exercises string `json:"exercises"`
	Concepts  string `json:"concepts"`
}

type TracksResponse struct {
	Tracks []Track `json:"tracks"`
}

type Exercise struct {
	Slug          string `json:"slug"`
	Type          string `json:"type"` // "tutorial", "concept", "practice"
	Title         string `json:"title"`
	IconURL       string `json:"icon_url"`
	Difficulty    string `json:"difficulty"` // "easy", "medium", "hard"
	Blurb         string `json:"blurb"`
	IsExternal    bool   `json:"is_external"`
	IsUnlocked    bool   `json:"is_unlocked"`
	IsRecommended bool   `json:"is_recommended"`
	Links         Links  `json:"links"`
}

type ExercisesResponse struct {
	Exercises []Exercise `json:"exercises"`
}

type CommunitySolution struct {
	UUID             string `json:"uuid"`
	NumStars         int    `json:"num_stars"`
	NumComments      int    `json:"num_comments"`
	NumLOC           int    `json:"num_loc"`
	PublishedAt      string `json:"published_at"`
	IsStarred        bool   `json:"is_starred"`
	PublishedIterTestsStatus string `json:"published_iteration_head_tests_status"`
	Author           struct {
		Handle    string `json:"handle"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
	Links struct {
		PublicURL string `json:"public_url"`
		Self      string `json:"self"`
	} `json:"links"`
}

type CommunitySolutionsResponse struct {
	Results []CommunitySolution `json:"results"`
	Meta    struct {
		CurrentPage  int `json:"current_page"`
		TotalCount   int `json:"total_count"`
		TotalPages   int `json:"total_pages"`
		UnscopedTotal int `json:"unscoped_total"`
	} `json:"meta"`
}

type CommunitySolutionFiles struct {
	Files []CommunitySolutionFile `json:"files"`
}

type CommunitySolutionFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Digest   string `json:"digest"`
}

type Solution struct {
	ID                  string         `json:"id"`
	URL                 string         `json:"url"`
	FileDownloadBaseURL string         `json:"file_download_base_url"`
	Files               []string       `json:"files"`
	Exercise            SolutionExInfo `json:"exercise"`
}

type SolutionExInfo struct {
	ID    string        `json:"id"`
	Track SolutionTrack `json:"track"`
}

type SolutionTrack struct {
	ID       string `json:"id"`
	Language string `json:"language"`
}

type SolutionResponse struct {
	Solution Solution `json:"solution"`
}
