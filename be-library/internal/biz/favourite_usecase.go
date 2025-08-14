package biz

import (
	"time"

)

type FavouriteUsecase struct {
	FavoriteModel data.FavoriteModel
	SeatTimeSlotsModel data.SeatTimeSlotsModel
	SeatsModel data.SeatsModel

	log      *log.Helper
}

func NewFavouriteUsecase(, ccnu CCNUServiceProxy, logger log.Logger, cf *conf.Server,
	que DelayQueue) LibraryUsecase {
	waitTime := 1200 * time.Millisecond

	if cf.Grpc.Timeout.Seconds > 0 {
		waitTime = cf.Grpc.Timeout.AsDuration()
	}

	uc := &libraryUsecase{
		crawler:  crawler,
		ccnu:     ccnu,
		log:      log.NewHelper(logger),
		waitTime: waitTime,
		Que:      que,
	}

	go func() {
		if err := uc.Que.Consume("be-library-refresh-retry", uc.handleRetryMsg); err != nil {
			uc.log.Errorf("Error consuming retry message: %v", err)
		}
	}()

	return uc
}


func (u *libraryUsecase) SeatToFavourite(seat *Seat, studentID string) *FavoriteSeat {
	return &FavoriteSeat{
		StudentID:  studentID,
		DevID:      seat.DevID,
		LabName:    seat.LabName,
		KindName:   seat.KindName,
		DevName:    seat.DevName,
		CreateTime: time.Now(),
	}
}

func (u *libraryUsecase) FavoriteToResponse(favourite *FavoriteSeat, includeTimeSlots bool) *FavoriteSeatResponse {
	response := &FavoriteSeatResponse{
		ID:          favorite.ID,
		DevID:       favorite.DevID,
		LabName:     favorite.LabName,
		KindName:    favorite.KindName,
		DevName:     favorite.DevName,
		IsAvailable: favorite.IsAvailable,
	}

	// 返回实时可用状态，去查询原始座位数据
	if includeTimeSlots {
		if seat := 
	}
}

func (u *libraryUsecase) AddFavourite(studentID string, seat *Seat) (*FavoriteSeat, error) {

}

// 设计为 输入 选中时间段，自动检测时间段是否可用吧？
func (U *libraryUsecase) calculateAvailability(timeSlot TimeSlot, devId_labId string) bool {
}
