package library

import (
	libraryv1 "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/asynccnu/ccnubox-be/bff/errs"
	"github.com/asynccnu/ccnubox-be/bff/pkg/ginx"
	"github.com/asynccnu/ccnubox-be/bff/web"
	"github.com/asynccnu/ccnubox-be/bff/web/ijwt"
	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	LibraryClient  libraryv1.LibraryClient // 注入 grpc 服务
	Administrators map[string]struct{}
}

func NewLibraryHandler(client libraryv1.LibraryClient, admins map[string]struct{}) *LibraryHandler {
	return &LibraryHandler{
		LibraryClient:  client,
		Administrators: admins,
	}
}

func (h *LibraryHandler) RegisterRoutes(s *gin.RouterGroup, authMiddlerware gin.HandlerFunc) {
	sg := s.Group("/library")
	sg.GET("/get_seat", authMiddlerware, ginx.WrapClaims(h.GetSeatInfos))
	sg.POST("/reserve_seat", authMiddlerware, ginx.WrapClaimsAndReq(h.ReserveSeat))
	sg.GET("/get_seat_records", authMiddlerware, ginx.WrapClaims(h.GetSeatRecord))
	sg.POST("/cancel_seat", authMiddlerware, ginx.WrapClaimsAndReq(h.CancelSeat))
	sg.GET("/get_credit_points", authMiddlerware, ginx.WrapClaims(h.GetCreditPoint))
	sg.POST("/get_discussion", authMiddlerware, ginx.WrapClaimsAndReq(h.GetDiscussion))
	sg.POST("/search_user", authMiddlerware, ginx.WrapClaimsAndReq(h.SearchUser))
	sg.POST("/reserve_discussion", authMiddlerware, ginx.WrapClaimsAndReq(h.ReserveDiscussion))
	sg.POST("/cancel_discussion", authMiddlerware, ginx.WrapClaimsAndReq(h.CancelDiscussion))
}

func (h *LibraryHandler) GetSeatInfos(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {

	res, err := h.LibraryClient.GetSeat(ctx, &libraryv1.GetSeatRequest{
		StuId: uc.StudentId,
	})

	if err != nil {
		return web.Response{}, errs.GET_SEAT_ERROR(err)
	}

	var roomList []Room

	for _, room := range res.RoomSeats {
		var seatList []Seat

		for _, seat := range room.Seats {
			var timeSlots []TimeSlot
			for _, ts := range seat.Ts {
				timeSlots = append(timeSlots, TimeSlot{
					Start:  ts.Start,
					End:    ts.End,
					State:  ts.State,
					Owner:  ts.Owner,
					Occupy: ts.Occupy,
				})
			}

			seatList = append(seatList, Seat{
				LabName:   seat.LabName,
				KindName:  seat.KindName,
				DevID:     seat.DevId,
				DevName:   seat.DevName,
				TimeSlots: timeSlots,
			})
		}

		roomList = append(roomList, Room{
			RoomID: room.RoomId,
			Seats:  seatList,
		})
	}

	resp := GetSeatResponse{
		Rooms: roomList,
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

func (h *LibraryHandler) ReserveSeat(ctx *gin.Context, req ReserveSeatRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.ReserveSeat(ctx, &libraryv1.ReserveSeatRequest{
		DevId: req.DevID,
		Start: req.Start,
		End:   req.End,
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.RESERVE_SEAT_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}

func (h *LibraryHandler) GetSeatRecord(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetSeatRecord(ctx, &libraryv1.GetSeatRecordRequest{
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_SEAT_RECORD_ERROR(err)
	}

	var respRecords = make([]Record, 0, len(res.Record))
	for _, record := range res.Record {
		respRecords = append(respRecords, Record{
			ID:       record.Id,
			Owner:    record.Owner,
			Start:    record.Start,
			End:      record.End,
			TimeDesc: record.TimeDesc,
			Occur:    record.Occur,
			States:   record.States,
			DevName:  record.DevName,
			RoomID:   record.RoomId,
			RoomName: record.RoomName,
			LabName:  record.LabName,
		})
	}

	resp := GetSeatRecordResponse{
		Records: respRecords,
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

func (h *LibraryHandler) CancelSeat(ctx *gin.Context, req CancelSeatRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.CancelSeat(ctx, &libraryv1.CancelSeatRequest{
		Id:    req.ID,
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.CANCEL_SEAT_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}

func (h *LibraryHandler) GetCreditPoint(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetCreditPoint(ctx, &libraryv1.GetCreditPointRequest{
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_CREDIT_POINTS_ERROR(err)
	}

	return web.Response{
		Msg:  "Success",
		Data: res,
	}, nil
}

func (h *LibraryHandler) GetDiscussion(ctx *gin.Context, req GetDiscussionRequest, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetDiscussion(ctx, &libraryv1.GetDiscussionRequest{
		ClassId: req.ClassID,
		Date:    req.Date,
		StuId:   uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_DISCUSSION_ERROR(err)
	}

	return web.Response{
		Msg:  "Success",
		Data: res,
	}, nil
}

func (h *LibraryHandler) SearchUser(ctx *gin.Context, req SearchUserRequest, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.SearchUser(ctx, &libraryv1.SearchUserRequest{
		StudentId: req.StudentID,
		StuId:     uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.SEARCH_USER_ERROR(err)
	}

	return web.Response{
		Msg:  "Success",
		Data: res,
	}, nil
}

func (h *LibraryHandler) ReserveDiscussion(ctx *gin.Context, req ReserveDiscussionRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.ReserveDiscussion(ctx, &libraryv1.ReserveDiscussionRequest{
		DevId:  req.DevID,
		LabId:  req.LabID,
		KindId: req.KindID,
		Title:  req.Title,
		Start:  req.Start,
		End:    req.End,
		List:   req.List,
		StuId:  uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.RESERVE_DISCUSSION_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}

func (h *LibraryHandler) CancelDiscussion(ctx *gin.Context, req CancelDiscussionRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.CancelDiscussion(ctx, &libraryv1.CancelDiscussionRequest{
		Id:    req.ID,
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.CANCEL_DISCUSSION_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}
