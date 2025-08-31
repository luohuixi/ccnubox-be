package biz

import "github.com/go-kratos/kratos/v2/log"

type FavouriteUsecase struct {
	FavoriteRepo FavoriteRepo
	SeatRepo     SeatRepo

	log *log.Helper
}

func NewFavouriteUsecase(favouriteRepo FavoriteRepo, seatRepo SeatRepo, logger log.Logger) *FavouriteUsecase {
	uc := &FavouriteUsecase{
		FavoriteRepo: favouriteRepo,
		SeatRepo:     seatRepo,

		log: log.NewHelper(logger),
	}

	return uc
}

func (u *FavouriteUsecase) AddFavourite() () {}
func (u *FavouriteUsecase) GetFavourite() () {}


// func (u *libraryUsecase) SeatToFavourite(seat *Seat, studentID string) *FavoriteSeat {
// 	return &FavoriteSeat{
// 		StudentID:  studentID,
// 		DevID:      seat.DevID,
// 		LabName:    seat.LabName,
// 		KindName:   seat.KindName,
// 		DevName:    seat.DevName,
// 		CreateTime: time.Now(),
// 	}
// }

// func (u *libraryUsecase) FavoriteToResponse(favourite *FavoriteSeat, includeTimeSlots bool) *FavoriteSeatResponse {
// 	response := &FavoriteSeatResponse{
// 		ID:          favorite.ID,
// 		DevID:       favorite.DevID,
// 		LabName:     favorite.LabName,
// 		KindName:    favorite.KindName,
// 		DevName:     favorite.DevName,
// 		IsAvailable: favorite.IsAvailable,
// 	}

// 	// 返回实时可用状态，去查询原始座位数据
// 	if includeTimeSlots {
// 		if seat :=
// 	}
// }

// func (u *libraryUsecase) AddFavourite(studentID string, seat *Seat) (*FavoriteSeat, error) {

// }

// // 设计为 输入 选中时间段，自动检测时间段是否可用吧？
// func (U *libraryUsecase) calculateAvailability(timeSlot TimeSlot, devId_labId string) bool {
// }
