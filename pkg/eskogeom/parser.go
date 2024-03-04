package eskogeom

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var uniqueCompoundIndex = 0

type PointType uint8

const (
	Invalid PointType = ' '
	MoveTo  PointType = 'm'
	LineTo  PointType = 'l'
	QuadTo  PointType = 'q'
	CubicTo PointType = 'c'
)

func ParsePoints(tokens []string, index int) ([]Point, PointType, int, error) {
	var points []Point
	pType := Invalid

	for index < len(tokens) {
		if len(tokens[index]) == 1 {
			match, _ := regexp.MatchString("^[a-z]$", tokens[index])
			if match {
				// We've reached a command, so break
				pType = PointType(tokens[index][0])
				index++
				break
			}
		}
		if index+1 >= len(tokens) {
			return nil, Invalid, index, errors.New("insufficient tokens for point")
		}

		x, err := strconv.ParseFloat(tokens[index], 64)
		if err != nil {
			return nil, Invalid, index, err
		}

		y, err := strconv.ParseFloat(tokens[index+1], 64)
		if err != nil {
			return nil, Invalid, index, err
		}

		points = append(points, Point{X: x, Y: y})
		index += 2
	}

	if pType == Invalid {
		return nil, Invalid, index, errors.New("point is missing a type")
	}

	if len(points) == 0 {
		return nil, Invalid, index, errors.New("not enough values for point")
	}

	return points, pType, index, nil
}

func ParseSubPath(tokens []string, startPos int) (*Path, int, error) {
	if len(tokens)-startPos < 3 {
		return nil, startPos, errors.New("insufficient tokens for a valid sub path")
	}

	subPath := &Path{}
	var err error
	var points []Point
	var pType PointType

	// Parse the initial moveTo point
	points, pType, startPos, err = ParsePoints(tokens, startPos)
	if err != nil {
		return nil, startPos, err
	}
	if pType != MoveTo {
		return nil, startPos, errors.New("sub path must begin with a MoveTo point")
	}
	subPath.MoveTo = points[0]

	// Parse all points for the current sub path
	for startPos < len(tokens) {

		if tokens[startPos][0] == 'h' {
			subPath.Closed = true
			startPos++
			break
		}

		// Parse points using ParsePoints function
		points, pType, startPos, err = ParsePoints(tokens, startPos)
		if err != nil {
			return nil, startPos, err
		}

		switch pType {
		case MoveTo:
			return subPath, startPos - 3, nil
		case LineTo:
			subPath.Points = append(subPath.Points, points[0])
		case QuadTo:
			subPath.Points = append(subPath.Points, Quad{
				C: points[0],
				P: points[1],
			})
		case CubicTo:
			subPath.Points = append(subPath.Points, Cubic{
				C1: points[0],
				C2: points[1],
				P:  points[2],
			})
		default:
			return nil, startPos, fmt.Errorf("unknown point type: %s", pType)
		}
	}

	return subPath, startPos, nil
}

func ParseCompound(tokens []string) (*Compound, error) {
	compound := &Compound{Index: uniqueCompoundIndex}
	uniqueCompoundIndex++
	pos := 0
	for pos < len(tokens) {
		subPath, newPos, err := ParseSubPath(tokens, pos)
		if err != nil {
			return nil, err
		}
		compound.SubPaths = append(compound.SubPaths, *subPath)
		pos = newPos
	}
	return compound, nil
}
