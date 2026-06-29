package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FilenameTimestampRegex regular expression for timestamps in filenames
var FilenameTimestampRegex *regexp.Regexp

func init() {
	FilenameTimestampRegex = regexp.MustCompile("[0-9]{4}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]+")
}

// RegexEndsWith returns a regex pattern that matches strings ending with val.
// val is escaped so that metacharacters are treated as literals.
func RegexEndsWith(val string) string {
	return fmt.Sprintf("^.*(%s)$", regexp.QuoteMeta(val))
}

// RegexEndsWithBeforeExt returns a regex pattern that matches strings
// ending with val before the file extension.
// val is escaped so that metacharacters are treated as literals.
func RegexEndsWithBeforeExt(val string) string {
	return fmt.Sprintf("^.*(%s)\\..*$", regexp.QuoteMeta(val))
}

// RegexBeginsWith returns a regex pattern that matches strings beginning with val.
// val is escaped so that metacharacters are treated as literals.
func RegexBeginsWith(val string) string {
	return fmt.Sprintf("^(%s).*$", regexp.QuoteMeta(val))
}

// Size returns the directory size in Bytes
func Size(path string, regex string) (uint64, error) {
	var size uint64
	isDesire, compErr := regexp.Compile(regex)
	if compErr != nil {
		return 0, compErr
	}
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if matched := isDesire.MatchString(info.Name()); matched || len(regex) == 0 {
				size += uint64(info.Size())
			}
		}
		return nil
	})
	return size, err
}

// List returns the files
func List(path string, regex string) ([]os.FileInfo, error) {
	result := make([]os.FileInfo, 0)
	isDesire, compErr := regexp.Compile(regex)
	if compErr != nil {
		return nil, compErr
	}
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if matched := isDesire.MatchString(info.Name()); matched || len(regex) == 0 {
				result = append(result, info)
			}
		}
		return nil
	})
	return result, err
}

// Expired returns the files that have expired
func Expired(path string, regex string, nowTime time.Time, maxTime time.Duration) ([]os.FileInfo, error) {
	result := make([]os.FileInfo, 0)
	isDesire, compErr := regexp.Compile(regex)
	if compErr != nil {
		return nil, compErr
	}
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if matched := isDesire.MatchString(info.Name()); matched || len(regex) == 0 {
				delta := nowTime.Sub(info.ModTime())
				if delta > maxTime {
					result = append(result, info)
				}
			}
		}
		return nil
	})
	return result, err
}

// BytesToMegaBytes converts Bytes to MegaBytes (SI decimal, 1 MB = 1,000,000 Bytes).
// For binary units (1 MiB = 1,048,576 Bytes), use BytesToMebiBytes.
func BytesToMegaBytes(in uint64) float64 {
	return float64(in) / 1000 / 1000
}

// BytesToGigaBytes converts Bytes to GigaBytes (SI decimal, 1 GB = 1,000,000,000 Bytes).
// For binary units (1 GiB = 1,073,741,824 Bytes), use BytesToGibiBytes.
func BytesToGigaBytes(in uint64) float64 {
	return float64(in) / 1000 / 1000 / 1000
}

// BytesToMebiBytes converts Bytes to MebiBytes (binary, 1 MiB = 1,048,576 Bytes).
func BytesToMebiBytes(in uint64) float64 {
	return float64(in) / 1024 / 1024
}

// BytesToGibiBytes converts Bytes to GibiBytes (binary, 1 GiB = 1,073,741,824 Bytes).
func BytesToGibiBytes(in uint64) float64 {
	return float64(in) / 1024 / 1024 / 1024
}

// AscendingTime sorts []os.FileInfo by modification time ascending,
// falling back to filename ascending if modification times are equal.
type AscendingTime []os.FileInfo

func (a AscendingTime) Len() int { return len(a) }
func (a AscendingTime) Less(i, j int) bool {
	ti, tj := a[i].ModTime(), a[j].ModTime()
	if ti.Equal(tj) {
		return a[i].Name() < a[j].Name()
	}
	return ti.Before(tj)
}
func (a AscendingTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// DescendingTime sorts []os.FileInfo by modification time descending,
// falling back to filename descending if modification times are equal.
type DescendingTime []os.FileInfo

func (a DescendingTime) Len() int { return len(a) }
func (a DescendingTime) Less(i, j int) bool {
	ti, tj := a[i].ModTime(), a[j].ModTime()
	if ti.Equal(tj) {
		return a[i].Name() > a[j].Name()
	}
	return ti.After(tj)
}
func (a DescendingTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// DescendingTimeName sorts []string by timestamp embedded in the name, descending.
type DescendingTimeName []string

func (a DescendingTimeName) Len() int { return len(a) }
func (a DescendingTimeName) Less(i, j int) bool {
	first := normalizeTimestamp(FilenameTimestampRegex.FindString(a[i]))
	second := normalizeTimestamp(FilenameTimestampRegex.FindString(a[j]))
	if first == second {
		return a[i] > a[j] // Tie-breaker
	}
	return first > second
}
func (a DescendingTimeName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// AscendingTimeName sorts []string by timestamp embedded in the name, ascending.
type AscendingTimeName []string

func (a AscendingTimeName) Len() int { return len(a) }
func (a AscendingTimeName) Less(i, j int) bool {
	first := normalizeTimestamp(FilenameTimestampRegex.FindString(a[i]))
	second := normalizeTimestamp(FilenameTimestampRegex.FindString(a[j]))
	if first == second {
		return a[i] < a[j] // Tie-breaker
	}
	return first < second
}
func (a AscendingTimeName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// normalizeTimestamp pads the subsecond field of a timestamp string to a fixed
// width so that lexicographic comparison produces correct chronological ordering.
// Expected input format: YYYY_MM_DD_HH_MM_SS_subseconds
func normalizeTimestamp(ts string) string {
	if ts == "" {
		return ts
	}
	const padWidth = 20
	idx := strings.LastIndex(ts, "_")
	if idx < 0 {
		return ts
	}
	prefix := ts[:idx+1]
	suffix := ts[idx+1:]
	if len(suffix) < padWidth {
		suffix = strings.Repeat("0", padWidth-len(suffix)) + suffix
	}
	return prefix + suffix
}
