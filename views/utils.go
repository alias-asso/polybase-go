package views

import (
	"fmt"
	"math/rand"
	"slices"
	"sort"
	"time"

	"github.com/alias-asso/polybase-go/libpolybase"
)

// SemesterGroup represents a group of courses for a semester
type SemesterGroup struct {
	Name    string
	Kinds   []KindGroup
	KindMap map[string]int
}

// KindGroup represents a group of courses of the same kind
type KindGroup struct {
	Name    string
	Courses []libpolybase.Course
}

func GroupCoursesBySemesterAndKind(courses []libpolybase.Course) []SemesterGroup {
	// Step 1: Get unique sorted semesters
	semesterMap := make(map[string]bool)
	for _, course := range courses {
		semesterMap[course.Semester] = true
	}
	semesters := make([]string, 0, len(semesterMap))
	for sem := range semesterMap {
		semesters = append(semesters, sem)
	}
	// Sort semesters by number in descending order (S2 before S1)
	sort.Slice(semesters, func(i, j int) bool {
		var num1, num2 int
		fmt.Sscanf(semesters[i], "S%d", &num1)
		fmt.Sscanf(semesters[j], "S%d", &num2)
		return num1 > num2
	})

	// Step 2: Get unique sorted kinds
	kindMap := make(map[int]bool)
	for _, course := range courses {
		kindMap[course.Year] = true
	}
	kinds := make([]int, 0, len(kindMap))
	for kind := range kindMap {
		kinds = append(kinds, kind)
	}
	slices.Sort(kinds)

	// Step 3: Create the structured result
	result := make([]SemesterGroup, len(semesters))

	// Initialize the semester groups
	for i, semester := range semesters {
		result[i] = SemesterGroup{
			Name:    semester,
			Kinds:   make([]KindGroup, len(kinds)),
			KindMap: make(map[string]int),
		}
		// Initialize kind groups
		for j, kind := range kinds {
			result[i].Kinds[j] = KindGroup{
				Name:    GetYear(kind),
				Courses: make([]libpolybase.Course, 0),
			}
			result[i].KindMap[GetYear(kind)] = j
		}
	}

	// Group courses
	for _, course := range courses {
		semIdx := slices.IndexFunc(result, func(sg SemesterGroup) bool { return sg.Name == course.Semester })
		if semIdx != -1 {
			kindIdx := result[semIdx].KindMap[GetYear(course.Year)]
			result[semIdx].Kinds[kindIdx].Courses = append(
				result[semIdx].Kinds[kindIdx].Courses,
				course,
			)
		}
	}

	// Sort courses within each group
	for i := range result {
		for j := range result[i].Kinds {
			sort.Slice(result[i].Kinds[j].Courses, func(m, n int) bool {
				return result[i].Kinds[j].Courses[m].Code < result[i].Kinds[j].Courses[n].Code
			})
		}
	}

	return result
}

var niceMessages = []string{
	"Nous espérons que tu passes une belle journée.",
	"Nya~",
	"Tu es une personne formidable.",
	"Tu as un talent monumental dans la gestion des polys.",
	"Tu as manqué à Polybase !",
	"Tu es mon membre préféré (ne le dis à personne).",
	"Tu es une personne très... très.",
	":3",
}

// GetRandomMessage returns a random nice message
func GetRandomMessage() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return niceMessages[r.Intn(len(niceMessages))]
}

func GetYear(i int) string {
	if i < 4 {
		return fmt.Sprintf("L%d", i)
	}
	return fmt.Sprintf("M%d", i%3)
}

func contains(courses []libpolybase.CourseID, id libpolybase.CourseID) bool {
	return slices.Contains(courses, id)
}
