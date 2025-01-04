package daylight

import (
	"math"
	"time"
)

type SunriseSunsetProvider interface {
	GetSunriseSunset(latitude, longitude float64, date time.Time) (time.Duration, time.Duration)
	GetSunriseSunsetAtOffset(latitude, longitude float64, date time.Time, offsetSeconds int) (time.Duration, time.Duration)
}

type sunriseSunsetProvider struct{}

func (s sunriseSunsetProvider) GetSunriseSunset(latitude, longitude float64, date time.Time) (time.Duration, time.Duration) {
	now := time.Now()
	_, tzOffsetSeconds := now.Zone()
	return s.GetSunriseSunsetAtOffset(latitude, longitude, date, tzOffsetSeconds)
}

func (s sunriseSunsetProvider) GetSunriseSunsetAtOffset(latitude, longitude float64, date time.Time, offsetSeconds int) (time.Duration, time.Duration) {
	res := newSunSet(latitude, longitude, float64(offsetSeconds)/3600.0)
	res.setCurrentDate(date.Year(), int(date.Month()), date.Day())
	sunrise := res.calcSunrise()
	sunset := res.calcSunset()
	return time.Duration(int(60.0*sunrise)) * time.Second, time.Duration(int(60.0*sunset)) * time.Second
}

func NewSunriseSunsetProvider() SunriseSunsetProvider {
	return &sunriseSunsetProvider{}
}

// Source: https://github.com/buelowp/sunset/tree/master/src
// Adapted from C to Go by Kirill Frolov
// Original license GPL V2: https://github.com/buelowp/sunset/blob/master/LICENSE

var SunsetOfficial = 90.833    /**< Standard sun angle for sunset */
var SunsetNautical = 102.0     /**< Nautical sun angle for sunset */
var SunsetCivil = 96.0         /**< Civil sun angle for sunset */
var SunsetAstronomical = 108.0 /**< Astronomical sun angle for sunset */

type SunSet struct {
	latitude   float64
	longitude  float64
	julianDate float64
	tzOffset   float64
}

func newSunSet(lat, lon, tz float64) *SunSet {
	return &SunSet{
		latitude:   lat,
		longitude:  lon,
		julianDate: 0,
		tzOffset:   tz,
	}
}

/**
 * \fn void SunSet::setPosition(float64 lat, float64 lon, int tz)
 * \param lat float64 Latitude value
 * \param lon float64 Longitude value
 * \param tz Integer Timezone offset
 *
 * This will set the location the library uses for it's math. The
 * timezone is included in this as it's not valid to call
 * any of the calc functions until you have set a timezone.
 * It is possible to simply call setPosition one time, with a timezone
 * and not use the setTZOffset() function ever, if you never
 * change timezone values.
 *
 * This is the old version of the setPosition using an integer
 * timezone, and will not be deprecated. However, it is preferred to
 * use the float64 version going forward.
 */
func (s *SunSet) SetPosition(lat, lon, tz float64) {
	s.latitude = lat
	s.longitude = lon
	if tz >= -12 && tz <= 14 {
		s.tzOffset = tz
	} else {
		s.tzOffset = 0.0
	}
}

func degToRad(angleDeg float64) float64 {
	return math.Pi * angleDeg / 180.0
}

func radToDeg(angleRad float64) float64 {
	return 180.0 * angleRad / math.Pi
}

func calcMeanObliquityOfEcliptic(t float64) float64 {
	seconds := 21.448 - t*(46.8150+t*(0.00059-t*(0.001813)))
	e0 := 23.0 + (26.0+(seconds/60.0))/60.0
	return e0 // in degrees
}

func calcGeomMeanLongSun(t float64) float64 {
	if math.IsNaN(t) {
		return math.NaN()
	}
	L := 280.46646 + t*(36000.76983+0.0003032*t)
	return math.Mod(L, 360.0)
}

func calcObliquityCorrection(t float64) float64 {
	e0 := calcMeanObliquityOfEcliptic(t)
	omega := 125.04 - 1934.136*t
	e := e0 + 0.00256*math.Cos(degToRad(omega))
	return e // in degrees
}

func calcEccentricityEarthOrbit(t float64) float64 {
	e := 0.016708634 - t*(0.000042037+0.0000001267*t)
	return e // unitless
}

func calcGeomMeanAnomalySun(t float64) float64 {
	M := 357.52911 + t*(35999.05029-0.0001537*t)
	return M // in degrees
}

func calcEquationOfTime(t float64) float64 {
	epsilon := calcObliquityCorrection(t)
	l0 := calcGeomMeanLongSun(t)
	e := calcEccentricityEarthOrbit(t)
	m := calcGeomMeanAnomalySun(t)
	y := math.Tan(degToRad(epsilon) / 2.0)

	y *= y

	sin2l0 := math.Sin(2.0 * degToRad(l0))
	sinm := math.Sin(degToRad(m))
	cos2l0 := math.Cos(2.0 * degToRad(l0))
	sin4l0 := math.Sin(4.0 * degToRad(l0))
	sin2m := math.Sin(2.0 * degToRad(m))
	Etime := y*sin2l0 - 2.0*e*sinm + 4.0*e*y*sinm*cos2l0 - 0.5*y*y*sin4l0 - 1.25*e*e*sin2m
	return radToDeg(Etime) * 4.0 // in minutes of time
}

func calcTimeJulianCent(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0
	return T
}

func calcSunTrueLong(t float64) float64 {
	l0 := calcGeomMeanLongSun(t)
	c := calcSunEqOfCenter(t)
	O := l0 + c
	return O // in degrees
}

func calcSunApparentLong(t float64) float64 {
	o := calcSunTrueLong(t)
	omega := 125.04 - 1934.136*t
	lambda := o - 0.00569 - 0.00478*math.Sin(degToRad(omega))
	return lambda // in degrees
}

func calcSunDeclination(t float64) float64 {
	e := calcObliquityCorrection(t)
	lambda := calcSunApparentLong(t)

	sint := math.Sin(degToRad(e)) * math.Sin(degToRad(lambda))
	theta := radToDeg(math.Asin(sint))
	return theta // in degrees
}

func calcHourAngleSunrise(lat, solarDec, offset float64) float64 {
	latRad := degToRad(lat)
	sdRad := degToRad(solarDec)
	HA := math.Acos(math.Cos(degToRad(offset))/(math.Cos(latRad)*math.Cos(sdRad)) - math.Tan(latRad)*math.Tan(sdRad))
	return HA // in radians
}

func calcHourAngleSunset(lat, solarDec, offset float64) float64 {
	latRad := degToRad(lat)
	sdRad := degToRad(solarDec)
	HA := math.Acos(math.Cos(degToRad(offset))/(math.Cos(latRad)*math.Cos(sdRad)) - math.Tan(latRad)*math.Tan(sdRad))
	return -HA // in radians
}

/**
 * \fn float64 SunSet::calcJD(int y, int m, int d) const
 * \param y Integer year as a 4 digit value
 * \param m Integer month, not 0 based
 * \param d Integer day, not 0 based
 * \return Returns the Julian date as a float64 for the calculations
 *
 * A well known JD calculator
 */
func calcJD(y, m, d int) float64 {
	if m <= 2 {
		y -= 1
		m += 12
	}
	A := math.Floor(float64(y) / 100)
	B := 2.0 - A + math.Floor(A/4)
	JD := math.Floor(365.25*(float64(y)+4716)) + math.Floor(30.6001*(float64(m)+1)) + float64(d) + B - 1524.5
	return JD
}

func calcJDFromJulianCent(t float64) float64 {
	JD := t*36525.0 + 2451545.0
	return JD
}

func calcSunEqOfCenter(t float64) float64 {
	m := calcGeomMeanAnomalySun(t)
	mrad := degToRad(m)
	sinm := math.Sin(mrad)
	sin2m := math.Sin(mrad + mrad)
	sin3m := math.Sin(mrad + mrad + mrad)
	C := sinm*(1.914602-t*(0.004817+0.000014*t)) + sin2m*(0.019993-0.000101*t) + sin3m*0.000289
	return C // in degrees
}

/**
 * \fn float64 SunSet::calcAbsSunrise(float64 offset) const
 * \param offset float64 The specific angle to use when calculating sunrise
 * \return Returns the time in minutes past midnight in UTC for sunrise at your location
 *
 * This does a bunch of work to get to an accurate angle. Note that it does it 2x, once
 * to get a rough position, and then it float64s back and redoes the calculations to
 * refine the value. The first time through, it will be off by as much as 2 minutes, but
 * the second time through, it will be nearly perfect.
 *
 * Note that this is the base calculation for all sunrise calls. The others just modify
 * the offset angle to account for the different needs.
 */
func (s *SunSet) calcAbsSunrise(offset float64) float64 {
	t := calcTimeJulianCent(s.julianDate)
	// *** First pass to approximate sunrise
	eqTime := calcEquationOfTime(t)
	solarDec := calcSunDeclination(t)
	hourAngle := calcHourAngleSunrise(s.latitude, solarDec, offset)
	delta := s.longitude + radToDeg(hourAngle)
	timeDiff := 4 * delta              // in minutes of time
	timeUTC := 720 - timeDiff - eqTime // in minutes
	newt := calcTimeJulianCent(calcJDFromJulianCent(t) + timeUTC/1440.0)

	eqTime = calcEquationOfTime(newt)
	solarDec = calcSunDeclination(newt)

	hourAngle = calcHourAngleSunrise(s.latitude, solarDec, offset)
	delta = s.longitude + radToDeg(hourAngle)
	timeDiff = 4 * delta
	timeUTC = 720 - timeDiff - eqTime // in minutes
	return timeUTC                    // return time in minutes from midnight
}

/**
 * \fn float64 SunSet::calcAbsSunset(float64 offset) const
 * \param offset float64 The specific angle to use when calculating sunset
 * \return Returns the time in minutes past midnight in UTC for sunset at your location
 *
 * This does a bunch of work to get to an accurate angle. Note that it does it 2x, once
 * to get a rough position, and then it float64s back and redoes the calculations to
 * refine the value. The first time through, it will be off by as much as 2 minutes, but
 * the second time through, it will be nearly perfect.
 *
 * Note that this is the base calculation for all sunset calls. The others just modify
 * the offset angle to account for the different needs.
 */
func (s *SunSet) calcAbsSunset(offset float64) float64 {
	t := calcTimeJulianCent(s.julianDate)
	// *** First pass to approximate sunset
	eqTime := calcEquationOfTime(t)
	solarDec := calcSunDeclination(t)
	hourAngle := calcHourAngleSunset(s.latitude, solarDec, offset)
	delta := s.longitude + radToDeg(hourAngle)
	timeDiff := 4 * delta              // in minutes of time
	timeUTC := 720 - timeDiff - eqTime // in minutes
	newt := calcTimeJulianCent(calcJDFromJulianCent(t) + timeUTC/1440.0)

	eqTime = calcEquationOfTime(newt)
	solarDec = calcSunDeclination(newt)

	hourAngle = calcHourAngleSunset(s.latitude, solarDec, offset)
	delta = s.longitude + radToDeg(hourAngle)
	timeDiff = 4 * delta
	timeUTC = 720 - timeDiff - eqTime // in minutes
	return timeUTC                    // return time in minutes from midnight
}

/**
 * \fn float64 SunSet::calcSunriseUTC()
 * \return Returns the UTC time when sunrise occurs in the location provided
 *
 * This is a holdover from the original implementation and to me doesn't
 * seem to be very useful, it's just confusing. This function is deprecated
 * but won't be removed unless that becomes necessary.
 */
func (s *SunSet) calcSunriseUTC() float64 {
	return s.calcAbsSunrise(SunsetOfficial)
}

/**
 * \fn float64 SunSet::calcSunsetUTC() const
 * \return Returns the UTC time when sunset occurs in the location provided
 *
 * This is a holdover from the original implementation and to me doesn't
 * seem to be very useful, it's just confusing. This function is deprecated
 * but won't be removed unless that becomes necessary.
 */
func (s *SunSet) calcSunsetUTC() float64 {
	return s.calcAbsSunset(SunsetOfficial)
}

/**
 * \fn float64 SunSet::calcAstronomicalSunrise()
 * \return Returns the Astronomical sunrise in fractional minutes past midnight
 *
 * This function will return the Astronomical sunrise in local time for your location
 */
func (s *SunSet) calcAstronomicalSunrise() float64 {
	return s.calcCustomSunrise(SunsetAstronomical)
}

/**
 * \fn float64 SunSet::calcAstronomicalSunset() const
 * \return Returns the Astronomical sunset in fractional minutes past midnight
 *
 * This function will return the Astronomical sunset in local time for your location
 */
func (s *SunSet) calcAstronomicalSunset() float64 {
	return s.calcCustomSunset(SunsetAstronomical)
}

/**
 * \fn float64 SunSet::calcCivilSunrise() const
 * \return Returns the Civil sunrise in fractional minutes past midnight
 *
 * This function will return the Civil sunrise in local time for your location
 */
func (s *SunSet) calcCivilSunrise() float64 {
	return s.calcCustomSunrise(SunsetCivil)
}

/**
 * \fn float64 SunSet::calcCivilSunset() const
 * \return Returns the Civil sunset in fractional minutes past midnight
 *
 * This function will return the Civil sunset in local time for your location
 */
func (s *SunSet) calcCivilSunset() float64 {
	return s.calcCustomSunset(SunsetCivil)
}

/**
 * \fn float64 SunSet::calcNauticalSunrise() const
 * \return Returns the Nautical sunrise in fractional minutes past midnight
 *
 * This function will return the Nautical sunrise in local time for your location
 */
func (s *SunSet) calcNauticalSunrise() float64 {
	return s.calcCustomSunrise(SunsetNautical)
}

/**
 * \fn float64 SunSet::calcNauticalSunset() const
 * \return Returns the Nautical sunset in fractional minutes past midnight
 *
 * This function will return the Nautical sunset in local time for your location
 */
func (s *SunSet) calcNauticalSunset() float64 {
	return s.calcCustomSunset(SunsetNautical)
}

/**
 * \fn float64 SunSet::calcSunrise() const
 * \return Returns local sunrise in minutes past midnight.
 *
 * This function will return the Official sunrise in local time for your location
 */
func (s *SunSet) calcSunrise() float64 {
	return s.calcCustomSunrise(SunsetOfficial)
}

/**
 * \fn float64 SunSet::calcSunset() const
 * \return Returns local sunset in minutes past midnight.
 *
 * This function will return the Official sunset in local time for your location
 */
func (s *SunSet) calcSunset() float64 {
	return s.calcCustomSunset(SunsetOfficial)
}

/**
 * \fn float64 SunSet::calcCustomSunrise(float64 angle) const
 * \param angle The angle in degrees over the horizon at which to calculate the sunset time
 * \return Returns sunrise at angle degrees in minutes past midnight.
 *
 * This function will return the sunrise in local time for your location for any
 * angle over the horizon, where < 90 would be above the horizon, and > 90 would be at or below.
 */
func (s *SunSet) calcCustomSunrise(angle float64) float64 {
	return s.calcAbsSunrise(angle) + (60 * s.tzOffset)
}

/**
 * \fn float64 SunSet::calcCustomSunset(float64 angle) const
 * \param angle The angle in degrees over the horizon at which to calculate the sunset time
 * \return Returns sunset at angle degrees in minutes past midnight.
 *
 * This function will return the sunset in local time for your location for any
 * angle over the horizon, where < 90 would be above the horizon, and > 90 would be at or below.
 */
func (s *SunSet) calcCustomSunset(angle float64) float64 {
	return s.calcAbsSunset(angle) + (60 * s.tzOffset)
}

/**
 * float64 SunSet::setCurrentDate(int y, int m, int d)
 * \param y Integer year, must be 4 digits
 * \param m Integer month, not zero based (Jan = 1)
 * \param d Integer day of month, not zero based (month starts on day 1)
 * \return Returns the result of the Julian Date conversion if you want to save it
 *
 * Since these calculations are done based on the Julian Calendar, we must convert
 * our year month day into Julian before we use it. You get the Julian value for
 * free if you want it.
 */
func (s *SunSet) setCurrentDate(y, m, d int) float64 {
	s.julianDate = calcJD(y, m, d)
	return s.julianDate
}

/**
 * \fn void SunSet::setTZOffset(int tz)
 * \param tz Integer timezone, may be positive or negative
 *
 * Critical to set your timezone so results are accurate for your time and date.
 * This function is critical to make sure the system works correctly. If you
 * do not set the timezone correctly, the return value will not be correct for
 * your location. Forgetting this will result in return values that may actually
 * be negative in some cases.
 *
 * This function is a holdover from the previous design using an integer timezone
 * and will not be deprecated. It is preferred to use the setTZOffset(doubble).
 */
func (s *SunSet) setTZOffsetInt(tz int) {
	if tz >= -12 && tz <= 14 {
		s.tzOffset = float64(tz)
	} else {
		s.tzOffset = 0.0
	}
}

/**
 * \fn void SunSet::setTZOffset(float64 tz)
 * \param tz float64 timezone, may be positive or negative
 *
 * Critical to set your timezone so results are accurate for your time and date.
 * This function is critical to make sure the system works correctly. If you
 * do not set the timezone correctly, the return value will not be correct for
 * your location. Forgetting this will result in return values that may actually
 * be negative in some cases.
 */
func (s *SunSet) setTZOffset(tz float64) {
	if tz >= -12 && tz <= 14 {
		s.tzOffset = tz
	} else {
		s.tzOffset = 0.0
	}
}
