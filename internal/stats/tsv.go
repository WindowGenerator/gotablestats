package stats

// TSVReader implements TableReader for TSV files
type TSVReader struct {
	*CSVReader
}

func NewTSVReader() *TSVReader {
	return &TSVReader{
		CSVReader: NewCSVReader('\t'),
	}
}

func (r *TSVReader) GetFormatName() string {
	return "TSV"
}
