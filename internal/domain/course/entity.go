package course

// Course 课程实体
type Course struct {
	SubID         int
	CourseID      int
	Name          string
	Teacher       string
	Location      string
	Date          string
	Time          string
	Video         string
	Asr           string
	SummaryStatus string
	SummaryData   string
	Model         string
	Token         uint32
	SummaryUser   string
}

// HasVideo 检查是否有视频
func (c *Course) HasVideo() bool {
	return c.Video != ""
}

// HasAsr 检查是否有ASR文本
func (c *Course) HasAsr() bool {
	return c.Asr != ""
}

// IsSummaryGenerating 检查摘要是否正在生成
func (c *Course) IsSummaryGenerating() bool {
	return c.SummaryStatus == "generating"
}

// IsSummaryFinished 检查摘要是否已完成
func (c *Course) IsSummaryFinished() bool {
	return c.SummaryStatus == "finished"
}
