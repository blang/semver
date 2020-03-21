package semver

// Equals checks if v is equal to o.
func (v Version) Equals(o Version) bool {
	return v.Compare(o) == 0
}

// EQ checks if v is equal to o.
func (v Version) EQ(o Version) bool {
	return v.Equals(o)
}

// NE checks if v is not equal to o.
func (v Version) NE(o Version) bool {
	return v.Compare(o) != 0
}

// GT checks if v is greater than o.
func (v Version) GT(o Version) bool {
	return v.Compare(o) == 1
}

// GTE checks if v is greater than or equal to o.
func (v Version) GTE(o Version) bool {
	return v.Compare(o) >= 0
}

// GE checks if v is greater than or equal to o.
func (v Version) GE(o Version) bool {
	return v.Compare(o) >= 0
}

// LT checks if v is less than o.
func (v Version) LT(o Version) bool {
	return v.Compare(o) == -1
}

// LTE checks if v is less than or equal to o.
func (v Version) LTE(o Version) bool {
	return v.Compare(o) <= 0
}

// LE checks if v is less than or equal to o.
func (v Version) LE(o Version) bool {
	return v.Compare(o) <= 0
}

// Compare compares Versions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v Version) Compare(o Version) int {
	if v.major != o.major {
		if v.major > o.major {
			return 1
		}
		return -1
	}
	if v.minor != o.minor {
		if v.minor > o.minor {
			return 1
		}
		return -1
	}
	if v.patch != o.patch {
		if v.patch > o.patch {
			return 1
		}
		return -1
	}

	// Quick comparison if a version has no prerelease versions
	if len(v.pre) == 0 && len(o.pre) == 0 {
		return 0
	} else if len(v.pre) == 0 && len(o.pre) > 0 {
		return 1
	} else if len(v.pre) > 0 && len(o.pre) == 0 {
		return -1
	}

	i := 0
	for ; i < len(v.pre) && i < len(o.pre); i++ {
		if comp := v.pre[i].Compare(o.pre[i]); comp == 0 {
			continue
		} else if comp == 1 {
			return 1
		} else {
			return -1
		}
	}

	// If all pr versions are the equal but one has further prerelease version, this one greater
	if i == len(v.pre) && i == len(o.pre) {
		return 0
	} else if i == len(v.pre) && i < len(o.pre) {
		return -1
	} else {
		return 1
	}
}

// Compare compares two PreRelease Versions v and o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v PRVersion) Compare(o PRVersion) int {
	if v.IsNum && !o.IsNum {
		return -1
	} else if !v.IsNum && o.IsNum {
		return 1
	} else if v.IsNum && o.IsNum {
		if v.VersionNum == o.VersionNum {
			return 0
		} else if v.VersionNum > o.VersionNum {
			return 1
		} else {
			return -1
		}
	} else { // both are Alphas
		if v.VersionStr == o.VersionStr {
			return 0
		} else if v.VersionStr > o.VersionStr {
			return 1
		} else {
			return -1
		}
	}
}
