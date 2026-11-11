package service

import (
	"errors"
	"fmt"
	"strconv"

	"backend-go/api-student/app/model"
)

var errUnknownEvent = errors.New("jenis event tidak dikenal")

func BuildMessage(event model.EnrollmentEvent, eventType string) (model.Message, error) {
	switch eventType {
	case model.EventEnrollmentCreated:
		return model.Message{
			Subject: "KRS: berhasil terdaftar di " + event.CourseCode + " " + event.CourseName,
			Body:    "Anda berhasil terdaftar di kelas " + event.CourseCode + " " + event.CourseName + " melalui KRS.",
		}, nil
	case model.EventEnrollmentDropped:
		return model.Message{
			Subject: "KRS: keluar dari " + event.CourseCode + " " + event.CourseName,
			Body:    "Anda telah keluar dari kelas " + event.CourseCode + " " + event.CourseName + " melalui KRS.",
		}, nil
	case model.EventEnrollmentSwapped:
		return model.Message{
			Subject: "KRS: pindah dari " + event.FromCourseCode + " " + event.FromCourseName +
				" ke " + event.CourseCode + " " + event.CourseName,
			Body: "KRS Anda dipindahkan dari kelas " + event.FromCourseCode + " " + event.FromCourseName +
				" ke kelas " + event.CourseCode + " " + event.CourseName + ".",
		}, nil
	case model.EventQuotaChanged:
		return model.Message{
			Subject: "KRS: kuota " + event.CourseCode + " " + event.CourseName +
				" berubah menjadi " + strconv.Itoa(event.Quota),
			Body: "Kuota kelas " + event.CourseCode + " " + event.CourseName +
				" kini menjadi " + strconv.Itoa(event.Quota) + " kursi.",
		}, nil
	}

	return model.Message{}, fmt.Errorf("%w: %q", errUnknownEvent, eventType)
}
