package dialog

func clampModalDimension(available, current, minimum, frame int) int {
	if current <= 0 {
		current = minimum
	}
	if available <= 0 {
		return current
	}

	maxSize := available - frame
	if maxSize < 1 {
		maxSize = 1
	}

	size := current
	if size > maxSize {
		size = maxSize
	}
	if size < minimum {
		if maxSize < minimum {
			size = maxSize
		} else {
			size = minimum
		}
	}

	return size
}

func clampModalWidth(availableWidth, currentWidth, minimumWidth int) int {
	return clampModalDimension(availableWidth, currentWidth, minimumWidth, 8)
}

func clampModalHeight(availableHeight, currentHeight, minimumHeight int) int {
	return clampModalDimension(availableHeight, currentHeight, minimumHeight, 4)
}
