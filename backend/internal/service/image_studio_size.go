package service

import "strings"

var ImageStudioRatios = []string{"1:1", "3:2", "2:3", "4:3", "3:4", "5:4", "4:5", "16:9", "9:16", "21:9"}
var ImageStudioResolutions = []string{"1K", "2K", "4K"}

var imageStudioSizeTable = map[string]map[string]string{
	"1K": {
		"1:1":  "1024x1024",
		"3:2":  "1216x832",
		"2:3":  "832x1216",
		"4:3":  "1152x864",
		"3:4":  "864x1152",
		"5:4":  "1120x896",
		"4:5":  "896x1120",
		"16:9": "1344x768",
		"9:16": "768x1344",
		"21:9": "1536x640",
	},
	"2K": {
		"1:1":  "1248x1248",
		"3:2":  "1536x1024",
		"2:3":  "1024x1536",
		"4:3":  "1440x1088",
		"3:4":  "1088x1440",
		"5:4":  "1392x1120",
		"4:5":  "1120x1392",
		"16:9": "1664x928",
		"9:16": "928x1664",
		"21:9": "1904x816",
	},
	"4K": {
		"1:1":  "2480x2480",
		"3:2":  "3056x2032",
		"2:3":  "2032x3056",
		"4:3":  "2880x2160",
		"3:4":  "2160x2880",
		"5:4":  "2784x2224",
		"4:5":  "2224x2784",
		"16:9": "3312x1872",
		"9:16": "1872x3312",
		"21:9": "3808x1632",
	},
}

var imageStudioSizeToTier = buildImageStudioSizeToTier()

func buildImageStudioSizeToTier() map[string]string {
	out := make(map[string]string, 32)
	for tier, ratios := range imageStudioSizeTable {
		for _, size := range ratios {
			out[strings.ToLower(size)] = tier
		}
	}
	return out
}

func ResolveImageStudioSize(ratio string, resolution string) (normalizedRatio string, normalizedResolution string, size string) {
	normalizedRatio = normalizeImageStudioRatio(ratio)
	normalizedResolution = normalizeImageStudioResolution(resolution)
	if ratios := imageStudioSizeTable[normalizedResolution]; ratios != nil {
		if size = ratios[normalizedRatio]; size != "" {
			return normalizedRatio, normalizedResolution, size
		}
		size = ratios["1:1"]
	}
	if size == "" {
		size = imageStudioSizeTable["1K"]["1:1"]
	}
	return normalizedRatio, normalizedResolution, size
}

func normalizeImageStudioRatio(ratio string) string {
	r := strings.TrimSpace(ratio)
	for _, allowed := range ImageStudioRatios {
		if r == allowed {
			return allowed
		}
	}
	return "1:1"
}

func normalizeImageStudioResolution(resolution string) string {
	r := strings.ToUpper(strings.TrimSpace(resolution))
	for _, allowed := range ImageStudioResolutions {
		if r == allowed {
			return allowed
		}
	}
	return "1K"
}

func openAIImageKnownSizeTier(size string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(size))
	if tier, ok := imageStudioSizeToTier[normalized]; ok {
		return tier, true
	}
	switch normalized {
	case "1792x1024", "1024x1792", "2048x2048", "2048x1152", "1152x2048":
		return "2K", true
	case "3840x2160", "2160x3840":
		return "4K", true
	default:
		return "", false
	}
}
