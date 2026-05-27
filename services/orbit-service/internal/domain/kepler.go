package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	muEarth      = 398600.4418 // km^3/s^2
	earthRadius  = 6378.137   // km
	flattening   = 1.0 / 298.257223563
	deg2rad      = math.Pi / 180.0
	rad2deg      = 180.0 / math.Pi
	secPerDay    = 86400.0
)

type keplerElements struct {
	epoch   time.Time
	incl    float64
	raan    float64
	ecc     float64
	argp    float64
	meanAnom float64
	meanMotion float64 // rad/s
	semiMajor  float64 // km
}

type KeplerPropagator struct{}

func NewKeplerPropagator() *KeplerPropagator {
	return &KeplerPropagator{}
}

func (p *KeplerPropagator) Propagate(tle TLE, start, end time.Time, step time.Duration) ([]GeodeticSample, error) {
	el, err := parseTLE(tle)
	if err != nil {
		return nil, err
	}
	if step <= 0 {
		step = 60 * time.Second
	}
	if !end.After(start) {
		return nil, fmt.Errorf("end must be after start")
	}

	var out []GeodeticSample
	for t := start; !t.After(end); t = t.Add(step) {
		lat, lon, alt, err := el.geodeticAt(t)
		if err != nil {
			return nil, err
		}
		out = append(out, GeodeticSample{
			Time:       t.UTC(),
			Latitude:   lat,
			Longitude:  lon,
			AltitudeKM: alt,
		})
	}
	for i := range out {
		out[i].Latitude = ClampLatDeg(out[i].Latitude)
		out[i].Longitude = WrapLonDeg(out[i].Longitude)
	}
	return out, nil
}

func (p *KeplerPropagator) FindPasses(tle TLE, station GroundStation, start, end time.Time, step time.Duration) ([]PassWindow, error) {
	if step <= 0 {
		step = 30 * time.Second
	}
	samples, err := p.Propagate(tle, start, end, step)
	if err != nil {
		return nil, err
	}

	minEl := station.MinElevationDeg
	if minEl <= 0 {
		minEl = 5
	}

	var passes []PassWindow
	inPass := false
	var cur PassWindow

	for _, s := range samples {
		el := elevationDeg(station.LatitudeDeg, station.LongitudeDeg, station.AltitudeM/1000.0, s.Latitude, s.Longitude, s.AltitudeKM)
		visible := el >= minEl
		if visible && !inPass {
			inPass = true
			cur = PassWindow{AOS: s.Time, MaxElevationDeg: el, MaxElevationTime: s.Time}
		} else if visible && inPass {
			if el > cur.MaxElevationDeg {
				cur.MaxElevationDeg = el
				cur.MaxElevationTime = s.Time
			}
			cur.LOS = s.Time
		} else if !visible && inPass {
			inPass = false
			if !cur.LOS.IsZero() && cur.LOS.After(cur.AOS) {
				passes = append(passes, cur)
			}
			cur = PassWindow{}
		}
	}
	if inPass && !cur.LOS.IsZero() && cur.LOS.After(cur.AOS) {
		passes = append(passes, cur)
	}
	return passes, nil
}

func parseTLE(tle TLE) (*keplerElements, error) {
	l1 := strings.TrimSpace(tle.Line1)
	l2 := strings.TrimSpace(tle.Line2)
	if len(l1) < 69 || len(l2) < 69 {
		return nil, fmt.Errorf("invalid TLE length")
	}

	epochYear, err := strconv.Atoi(strings.TrimSpace(l1[18:20]))
	if err != nil {
		return nil, fmt.Errorf("epoch year: %w", err)
	}
	epochDay, err := strconv.ParseFloat(strings.TrimSpace(l1[20:32]), 64)
	if err != nil {
		return nil, fmt.Errorf("epoch day: %w", err)
	}
	year := 2000 + epochYear
	if epochYear >= 57 {
		year = 1900 + epochYear
	}
	epoch := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration((epochDay - 1) * float64(time.Hour*24)))

	incl, _ := strconv.ParseFloat(strings.TrimSpace(l2[8:16]), 64)
	raan, _ := strconv.ParseFloat(strings.TrimSpace(l2[17:25]), 64)
	ecc, _ := strconv.ParseFloat("0."+strings.TrimSpace(l2[26:33]), 64)
	argp, _ := strconv.ParseFloat(strings.TrimSpace(l2[34:42]), 64)
	m0, _ := strconv.ParseFloat(strings.TrimSpace(l2[43:51]), 64)
	nRev, _ := strconv.ParseFloat(strings.TrimSpace(l2[52:63]), 64)

	n := nRev * 2 * math.Pi / secPerDay
	a := math.Cbrt(muEarth / (n * n))

	return &keplerElements{
		epoch:      epoch,
		incl:       incl * deg2rad,
		raan:       raan * deg2rad,
		ecc:        ecc,
		argp:       argp * deg2rad,
		meanAnom:   m0 * deg2rad,
		meanMotion: n,
		semiMajor:  a,
	}, nil
}

func (el *keplerElements) geodeticAt(t time.Time) (lat, lon, altKM float64, err error) {
	dt := t.Sub(el.epoch).Seconds()
	M := normalizeAngle(el.meanAnom + el.meanMotion*dt)
	E := solveKepler(M, el.ecc)
	nu := 2 * math.Atan2(math.Sqrt(1+el.ecc)*math.Sin(E/2), math.Sqrt(1-el.ecc)*math.Cos(E/2))
	r := el.semiMajor * (1 - el.ecc*math.Cos(E))

	xOrb := r * math.Cos(nu)
	yOrb := r * math.Sin(nu)

	cosO, sinO := math.Cos(el.raan), math.Sin(el.raan)
	cosI, sinI := math.Cos(el.incl), math.Sin(el.incl)
	cosW, sinW := math.Cos(el.argp), math.Sin(el.argp)

	x := (cosO*cosW-sinO*sinW*cosI)*xOrb + (-cosO*sinW-sinO*cosW*cosI)*yOrb
	y := (sinO*cosW+cosO*sinW*cosI)*xOrb + (-sinO*sinW+cosO*cosW*cosI)*yOrb
	z := (sinW*sinI)*xOrb + (cosW*sinI)*yOrb

	// rotate ECEF approx: ignore polar motion; apply GMST
	gmst := greenwichSiderealAngle(t)
	cosG, sinG := math.Cos(gmst), math.Sin(gmst)
	xe := cosG*x + sinG*y
	ye := -sinG*x + cosG*y
	ze := z

	return ecefToGeodetic(xe, ye, ze)
}

func solveKepler(M, e float64) float64 {
	E := M
	for i := 0; i < 12; i++ {
		d := E - e*math.Sin(E) - M
		E -= d / (1 - e*math.Cos(E))
		if math.Abs(d) < 1e-10 {
			break
		}
	}
	return E
}

func normalizeAngle(a float64) float64 {
	twoPi := 2 * math.Pi
	a = math.Mod(a, twoPi)
	if a < 0 {
		a += twoPi
	}
	return a
}

func greenwichSiderealAngle(t time.Time) float64 {
	jd := julianDate(t)
	tUT := (jd - 2451545.0) / 36525.0
	gmst := 280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*tUT*tUT
	return normalizeAngle(gmst * deg2rad)
}

func julianDate(t time.Time) float64 {
	unix := float64(t.UTC().UnixNano()) / 1e9
	return 2440587.5 + unix/86400.0
}

func ecefToGeodetic(x, y, z float64) (lat, lon, altKM float64, err error) {
	lon = math.Atan2(y, x) * rad2deg
	p := math.Hypot(x, y)
	lat = math.Atan2(z, p*(1-flattening))
	for i := 0; i < 5; i++ {
		sinLat := math.Sin(lat)
		N := earthRadius / math.Sqrt(1-flattening*(2-flattening)*sinLat*sinLat)
		altKM = p/math.Cos(lat) - N
		lat = math.Atan2(z, p*(1-flattening*(2-flattening)*N/(N+altKM)))
	}
	lat *= rad2deg
	if lon > 180 {
		lon -= 360
	}
	if lon < -180 {
		lon += 360
	}
	return lat, lon, altKM, nil
}

func elevationDeg(lat1, lon1, alt1KM, lat2, lon2, alt2KM float64) float64 {
	x1, y1, z1 := geodeticToECEF(lat1, lon1, alt1KM)
	x2, y2, z2 := geodeticToECEF(lat2, lon2, alt2KM)
	rx, ry, rz := x2-x1, y2-y1, z2-z1
	// topocentric ENU
	lat := lat1 * deg2rad
	lon := lon1 * deg2rad
	sinLat, cosLat := math.Sin(lat), math.Cos(lat)
	sinLon, cosLon := math.Sin(lon), math.Cos(lon)
	east := -sinLon*rx + cosLon*ry
	north := -sinLat*cosLon*rx - sinLat*sinLon*ry + cosLat*rz
	up := cosLat*cosLon*rx + cosLat*sinLon*ry + sinLat*rz
	horiz := math.Hypot(east, north)
	return math.Atan2(up, horiz) * rad2deg
}

func geodeticToECEF(latDeg, lonDeg, altKM float64) (x, y, z float64) {
	lat := latDeg * deg2rad
	lon := lonDeg * deg2rad
	sinLat, cosLat := math.Sin(lat), math.Cos(lat)
	sinLon, cosLon := math.Sin(lon), math.Cos(lon)
	N := earthRadius / math.Sqrt(1-flattening*(2-flattening)*sinLat*sinLat)
	x = (N + altKM) * cosLat * cosLon
	y = (N + altKM) * cosLat * sinLon
	z = (N*(1-flattening*(2-flattening)) + altKM) * sinLat
	return
}
