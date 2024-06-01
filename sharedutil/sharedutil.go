package sharedutil

import (
	"math"
	"slices"

	"github.com/dweymouth/supersonic/backend/mediaprovider"
)

func FilterSlice[T any](ss []T, test func(T) bool) []T {
	if ss == nil {
		return nil
	}
	result := make([]T, 0)
	for _, s := range ss {
		if test(s) {
			result = append(result, s)
		}
	}
	return result
}

func MapSlice[T any, U any](ts []T, f func(T) U) []U {
	if ts == nil {
		return nil
	}
	result := make([]U, len(ts))
	for i, t := range ts {
		result[i] = f(t)
	}
	return result
}

func FilterMapSlice[T any, U any](ts []T, f func(T) (U, bool)) []U {
	if ts == nil {
		return nil
	}
	result := make([]U, 0)
	for _, t := range ts {
		if u, ok := f(t); ok {
			result = append(result, u)
		}
	}
	return result
}

func Reversed[T any](ts []T) []T {
	if ts == nil {
		return nil
	}
	new := make([]T, len(ts))
	j := len(ts) - 1
	for i := range ts {
		new[i] = ts[j]
		j--
	}
	return new
}

func ToSet[T comparable](ts []T) map[T]struct{} {
	set := make(map[T]struct{}, len(ts))
	for _, t := range ts {
		set[t] = struct{}{}
	}
	return set
}

func FindTrackByID(id string, tracks []*mediaprovider.Track) *mediaprovider.Track {
	for _, tr := range tracks {
		if id == tr.ID {
			return tr
		}
	}
	return nil
}

func FindMediaItemByID(id string, items []mediaprovider.MediaItem) mediaprovider.MediaItem {
	for _, tr := range items {
		if id == tr.Metadata().ID {
			return tr
		}
	}
	return nil
}

func MediaItemIDOrEmptyStr(item mediaprovider.MediaItem) string {
	if tr, ok := item.(*mediaprovider.Track); ok && tr != nil {
		return tr.ID
	}
	if rd, ok := item.(*mediaprovider.RadioStation); ok && rd != nil {
		return rd.ID
	}
	return ""
}

func AlbumIDOrEmptyStr(track *mediaprovider.Track) string {
	if track == nil {
		return ""
	}
	return track.AlbumID
}

func TracksToIDs(tracks []*mediaprovider.Track) []string {
	return MapSlice(tracks, func(tr *mediaprovider.Track) string {
		return tr.ID
	})
}

type TrackReorderOp int

const (
	MoveToTop TrackReorderOp = iota
	MoveToBottom
	MoveUp
	MoveDown
)

// TODO: it's a shame the below function is just duplicated for a slice of mediaprovider.MediaItem
// Find out if there's a better way with refactoring.

// Reorder tracks and return a new track slice.
// idxToMove must contain only valid indexes into tracks, and no repeats
func ReorderTracks(tracks []*mediaprovider.Track, idxToMove []int, op TrackReorderOp) []*mediaprovider.Track {
	newTracks := make([]*mediaprovider.Track, len(tracks))
	switch op {
	case MoveToTop:
		topIdx := 0
		botIdx := len(idxToMove)
		idxToMoveSet := ToSet(idxToMove)
		for i, t := range tracks {
			if _, ok := idxToMoveSet[i]; ok {
				newTracks[topIdx] = t
				topIdx++
			} else {
				newTracks[botIdx] = t
				botIdx++
			}
		}
	case MoveToBottom:
		topIdx := 0
		botIdx := len(tracks) - len(idxToMove)
		idxToMoveSet := ToSet(idxToMove)
		for i, t := range tracks {
			if _, ok := idxToMoveSet[i]; ok {
				newTracks[botIdx] = t
				botIdx++
			} else {
				newTracks[topIdx] = t
				topIdx++
			}
		}
	case MoveUp:
		first := firstIdxCanMoveUp(idxToMove)
		copy(newTracks, tracks)
		for _, i := range idxToMove {
			if i < first {
				continue
			}
			newTracks[i-1], newTracks[i] = newTracks[i], newTracks[i-1]
		}
	case MoveDown:
		last := lastIdxCanMoveDown(idxToMove, len(tracks))
		copy(newTracks, tracks)
		for i := len(idxToMove) - 1; i >= 0; i-- {
			idx := idxToMove[i]
			if idx > last {
				continue
			}
			newTracks[idx+1], newTracks[idx] = newTracks[idx], newTracks[idx+1]
		}
	}
	return newTracks
}

// Reorder media items and return a new item slice.
// idxToMove must contain only valid indexes into items, and no repeats
func ReorderMediaItems(items []mediaprovider.MediaItem, idxToMove []int, op TrackReorderOp) []mediaprovider.MediaItem {
	newTracks := make([]mediaprovider.MediaItem, len(items))
	switch op {
	case MoveToTop:
		topIdx := 0
		botIdx := len(idxToMove)
		idxToMoveSet := ToSet(idxToMove)
		for i, t := range items {
			if _, ok := idxToMoveSet[i]; ok {
				newTracks[topIdx] = t
				topIdx++
			} else {
				newTracks[botIdx] = t
				botIdx++
			}
		}
	case MoveToBottom:
		topIdx := 0
		botIdx := len(items) - len(idxToMove)
		idxToMoveSet := ToSet(idxToMove)
		for i, t := range items {
			if _, ok := idxToMoveSet[i]; ok {
				newTracks[botIdx] = t
				botIdx++
			} else {
				newTracks[topIdx] = t
				topIdx++
			}
		}
	case MoveUp:
		first := firstIdxCanMoveUp(idxToMove)
		copy(newTracks, items)
		for _, i := range idxToMove {
			if i < first {
				continue
			}
			newTracks[i-1], newTracks[i] = newTracks[i], newTracks[i-1]
		}
	case MoveDown:
		last := lastIdxCanMoveDown(idxToMove, len(items))
		copy(newTracks, items)
		for i := len(idxToMove) - 1; i >= 0; i-- {
			idx := idxToMove[i]
			if idx > last {
				continue
			}
			newTracks[idx+1], newTracks[idx] = newTracks[idx], newTracks[idx+1]
		}
	}
	return newTracks
}

func firstIdxCanMoveUp(idxs []int) int {
	prevIdx := -1
	slices.Sort(idxs)
	for _, idx := range idxs {
		if idx > prevIdx+1 {
			return idx
		}
		prevIdx = idx
	}
	return math.MaxInt
}

func lastIdxCanMoveDown(idxs []int, lenSlice int) int {
	prevIdx := lenSlice
	slices.Sort(idxs)
	for i := len(idxs) - 1; i >= 0; i-- {
		idx := idxs[i]
		if idx < prevIdx-1 {
			return idx
		}
		prevIdx = idx
	}
	return -1
}
