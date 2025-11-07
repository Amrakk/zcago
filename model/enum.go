package model

type ThreadType uint8

const (
	ThreadTypeUser ThreadType = iota
	ThreadTypeGroup
)

type DestType uint8

const (
	DestTypeGroup DestType = 1
	DestTypeUser  DestType = 3
	DestTypePage  DestType = 5
)

type EventType int

const (
	EventTypeGroup EventType = iota
	EventTypeFriend
)

// BinBankCard represents the BIN codes of
// banks supported by Zalo, collected via MITM inspection on the mobile app.
//
// [Documentation]: docs missing bin code and short_name bank
//
// [Documentation]: https://developers.zalo.me/docs/zalo-notification-service/phu-luc/danh-sach-bin-code
type BinBankCard int

const (
	BinBankABBank                    BinBankCard = 970425  // NH TMCP An Bình
	BinBankACB                       BinBankCard = 970416  // NH TMCP Á Châu
	BinBankAgribank                  BinBankCard = 970405  // NH Nông nghiệp và Phát triển Nông thôn Việt Nam
	BinBankBIDV                      BinBankCard = 970418  // NH TMCP Đầu tư và Phát triển Việt Nam
	BinBankBVBank                    BinBankCard = 970454  // NH TMCP Bản Việt
	BinBankBacABank                  BinBankCard = 970409  // NH TMCP Bắc Á
	BinBankBaoVietBank               BinBankCard = 970438  // NH TMCP Bảo Việt
	BinBankCAKE                      BinBankCard = 546034  // NH số CAKE by VPBank
	BinBankCBBank                    BinBankCard = 970444  // NH Thương mại TNHH MTV Xây dựng Việt Nam
	BinBankCIMBBank                  BinBankCard = 422589  // NH TNHH MTV CIMB Việt Nam
	BinBankCoopBank                  BinBankCard = 970446  // NH Hợp tác xã Việt Nam
	BinBankDBSBank                   BinBankCard = 796500  // NH TNHH MTV Phát triển Singapore - CN TP. Hồ Chí Minh
	BinBankDongABank                 BinBankCard = 970406  // NH TMCP Đông Á
	BinBankEximbank                  BinBankCard = 970431  // NH TMCP Xuất Nhập khẩu Việt Nam
	BinBankGPBank                    BinBankCard = 970408  // NH TMCP Dầu khí Toàn cầu
	BinBankHDBank                    BinBankCard = 970437  // NH TMCP Phát triển TP. Hồ Chí Minh
	BinBankHSBC                      BinBankCard = 458761  // NH TNHH MTV HSBC (Việt Nam)
	BinBankHongLeongBank             BinBankCard = 970442  // NH TNHH MTV Hong Leong Việt Nam
	BinBankIBKHCM                    BinBankCard = 970456  // NH Công nghiệp Hàn Quốc - CN TP. Hồ Chí Minh
	BinBankIBKHN                     BinBankCard = 970455  // NH Công nghiệp Hàn Quốc - CN Hà Nội
	BinBankIndovinaBank              BinBankCard = 970434  // NH TNHH Indovina
	BinBankKBank                     BinBankCard = 668888  // NH Đại chúng TNHH Kasikornbank - CN TP. Hồ Chí Minh
	BinBankKienlongBank              BinBankCard = 970452  // NH TMCP Kiên Long
	BinBankKookminBankHCM            BinBankCard = 970463  // NH Kookmin - CN TP. Hồ Chí Minh
	BinBankKookminBankHN             BinBankCard = 970462  // NH Kookmin - CN Hà Nội
	BinBankLPBank                    BinBankCard = 970449  // NH TMCP Lộc Phát Việt Nam
	BinBankMBBank                    BinBankCard = 970422  // NH TMCP Quân đội
	BinBankMSB                       BinBankCard = 970426  // NH TMCP Hàng Hải
	BinBankNCB                       BinBankCard = 970419  // NH TMCP Quốc Dân
	BinBankNamABank                  BinBankCard = 970428  // NH TMCP Nam Á
	BinBankNongHyupBank              BinBankCard = 801011  // NH Nonghyup - CN Hà Nội
	BinBankOCB                       BinBankCard = 970448  // NH TMCP Phương Đông
	BinBankOceanBank                 BinBankCard = 970414  // NH Thương mại TNHH MTV Đại Dương
	BinBankPGBank                    BinBankCard = 970430  // NH TMCP Thịnh vượng và Phát triển
	BinBankPVcomBank                 BinBankCard = 970412  // NH TMCP Đại Chúng Việt Nam
	BinBankPublicBankVietnam         BinBankCard = 970439  // NH TNHH MTV Public Việt Nam
	BinBankSCB                       BinBankCard = 970429  // NH TMCP Sài Gòn
	BinBankSHB                       BinBankCard = 970443  // NH TMCP Sài Gòn - Hà Nội
	BinBankSacombank                 BinBankCard = 970403  // NH TMCP Sài Gòn Thương Tín
	BinBankSaigonBank                BinBankCard = 970400  // NH TMCP Sài Gòn Công Thương
	BinBankSeABank                   BinBankCard = 970440  // NH TMCP Đông Nam Á
	BinBankShinhanBank               BinBankCard = 970424  // NH TNHH MTV Shinhan Việt Nam
	BinBankStandardCharteredVietnam  BinBankCard = 970410  // NH TNHH MTV Standard Chartered Bank Việt Nam
	BinBankTNEX                      BinBankCard = 9704261 // NH số TNEX
	BinBankTPBank                    BinBankCard = 970423  // NH TMCP Tiên Phong
	BinBankTechcombank               BinBankCard = 970407  // NH TMCP Kỹ thương Việt Nam
	BinBankTimo                      BinBankCard = 963388  // NH số Timo by Bản Việt Bank
	BinBankUBank                     BinBankCard = 546035  // NH số UBank by VPBank
	BinBankUnitedOverseasBankVietnam BinBankCard = 970458  // NH United Overseas Bank Việt Nam
	BinBankVIB                       BinBankCard = 970441  // NH TMCP Quốc tế Việt Nam
	BinBankVPBank                    BinBankCard = 970432  // NH TMCP Việt Nam Thịnh Vượng
	BinBankVRB                       BinBankCard = 970421  // NH Liên doanh Việt - Nga
	BinBankVietABank                 BinBankCard = 970427  // NH TMCP Việt Á
	BinBankVietBank                  BinBankCard = 970433  // NH TMCP Việt Nam Thương Tín
	BinBankVietcombank               BinBankCard = 970436  // NH TMCP Ngoại Thương Việt Nam
	BinBankVietinBank                BinBankCard = 970415  // NH TMCP Công thương Việt Nam
	BinBankWooriBank                 BinBankCard = 970457  // NH TNHH MTV Woori Việt Nam
)
