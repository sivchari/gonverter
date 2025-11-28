package analyzer

// Convert_DifferentFields_Name_To_TargetDiff_Name is a custom conversion function
func Convert_DifferentFields_Name_To_TargetDiff_Name(src *DifferentFields, dst *TargetDiff) {
	dst.Name = src.FullName
}
