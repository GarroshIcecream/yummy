package dialog

func clampModalDimension(available, current, minimum, frame int) int {
	if current <= 0 {
		current = minimum
	}
	if available <= 0 {
		return current
	}

	maxSize := max(available-frame, 1)

	size := min(current, maxSize)
	if size < minimum {
		size = min(maxSize, minimum)
	}

	return size
}

func clampModalWidth(availableWidth, currentWidth, minimumWidth int) int {
	return clampModalDimension(availableWidth, currentWidth, minimumWidth, 8)
}

func clampModalHeight(availableHeight, currentHeight, minimumHeight int) int {
	return clampModalDimension(availableHeight, currentHeight, minimumHeight, 4)
}
