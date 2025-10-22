package model

type ZBusinessPackage struct {
	Label any   `json:"label"`
	PkgId uint8 `json:"pkgId"`
}

type BusinessCategory int

const (
	Other BusinessCategory = iota
	RealEstate
	TechnologyAndDevices
	TravelAndHospitality
	EducationAndTraining
	ShoppingAndRetail
	CosmeticsAndBeauty
	RestaurantAndCafe
	AutoAndMotorbike
	FashionAndApparel
	FoodAndBeverage
	MediaAndEntertainment
	InternalCommunications
	Transportation
	Telecommunications
)

var BusinessCategoryName = map[BusinessCategory]string{
	Other:                  "Dịch vụ khác (Không hiển thị)",
	RealEstate:             "Bất động sản",
	TechnologyAndDevices:   "Công nghệ & Thiết bị",
	TravelAndHospitality:   "Du lịch & Lưu trú",
	EducationAndTraining:   "Giáo dục & Đào tạo",
	ShoppingAndRetail:      "Mua sắm & Bán lẻ",
	CosmeticsAndBeauty:     "Mỹ phẩm & Làm đẹp",
	RestaurantAndCafe:      "Nhà hàng & Quán",
	AutoAndMotorbike:       "Ô tô & Xe máy",
	FashionAndApparel:      "Thời trang & May mặc",
	FoodAndBeverage:        "Thực phẩm & Đồ uống",
	MediaAndEntertainment:  "Truyền thông & Giải trí",
	InternalCommunications: "Truyền thông nội bộ",
	Transportation:         "Vận tải",
	Telecommunications:     "Viễn thông",
}
