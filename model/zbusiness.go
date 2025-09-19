package model

type ZBusinessPackage struct {
	Label *interface{} `json:"label"`
	PkgId uint8        `json:"pkgId"`
}

type BusinessCategory int

const (
	Other                  BusinessCategory = 0
	RealEstate             BusinessCategory = 1
	TechnologyAndDevices   BusinessCategory = 2
	TravelAndHospitality   BusinessCategory = 3
	EducationAndTraining   BusinessCategory = 4
	ShoppingAndRetail      BusinessCategory = 5
	CosmeticsAndBeauty     BusinessCategory = 6
	RestaurantAndCafe      BusinessCategory = 7
	AutoAndMotorbike       BusinessCategory = 8
	FashionAndApparel      BusinessCategory = 9
	FoodAndBeverage        BusinessCategory = 10
	MediaAndEntertainment  BusinessCategory = 11
	InternalCommunications BusinessCategory = 12
	Transportation         BusinessCategory = 13
	Telecommunications     BusinessCategory = 14
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
