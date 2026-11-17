package service

import (
	"errors"
	"fmt"
	"strconv"

	"backend-go/api-student/app/model"
)

var errUnknownEvent = errors.New("jenis event tidak dikenal")

func BuildMessage(eventType string, event model.EnrollmentEvent, user model.User) (model.Message, error) {
	switch eventType {
	case model.EventEnrollmentCreated:
		return model.Message{
			To:      user.Email,
			Subject: "KRS: berhasil terdaftar di " + event.CourseCode + " " + event.CourseName,
			Body: "Halo " + user.Username + ", Anda berhasil terdaftar di kelas " +
				event.CourseCode + " " + event.CourseName + " melalui KRS.",
		}, nil
	case model.EventEnrollmentDropped:
		return model.Message{
			To:      user.Email,
			Subject: "KRS: keluar dari " + event.CourseCode + " " + event.CourseName,
			Body: "Halo " + user.Username + ", Anda telah keluar dari kelas " +
				event.CourseCode + " " + event.CourseName + " melalui KRS.",
		}, nil
	case model.EventEnrollmentSwapped:
		return model.Message{
			To: user.Email,
			Subject: "KRS: pindah dari " + event.FromCourseCode + " " + event.FromCourseName +
				" ke " + event.CourseCode + " " + event.CourseName,
			Body: "Halo " + user.Username + ", KRS Anda dipindahkan dari kelas " +
				event.FromCourseCode + " " + event.FromCourseName +
				" ke kelas " + event.CourseCode + " " + event.CourseName + ".",
		}, nil
	case model.EventQuotaChanged:
		return model.Message{
			To: user.Email,
			Subject: "KRS: kuota " + event.CourseCode + " " + event.CourseName +
				" berubah menjadi " + strconv.Itoa(event.Quota),
			Body: "Halo " + user.Username + ", kuota kelas " + event.CourseCode + " " +
				event.CourseName + " kini menjadi " + strconv.Itoa(event.Quota) + " kursi.",
		}, nil
	}

	return model.Message{}, fmt.Errorf("%w: %q", errUnknownEvent, eventType)
}
