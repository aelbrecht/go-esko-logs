package eskolog

import (
	"fmt"
	"github.com/aelbrecht/go-esko-logs/pkg/eskogeom"
	"log"
	"sort"
	"strconv"
	"strings"
)

var DefaultGroupName = "Default"

var uniqueGroupIndex = 0

type Collection struct {
	Sessions []*Session `json:"sessions"`
}

type Group struct {
	Index      int                  `json:"index"`
	Name       string               `json:"name"`
	Compounds  []*eskogeom.Compound `json:"compounds"`
	Messages   []*LogEntry          `json:"messages"`
	Partitions int                  `json:"partitions"`
}

func (l *Group) AddCompound(compound *eskogeom.Compound) {
	l.Compounds = append(l.Compounds, compound)
}

func (l *Group) AddMessage(message *LogEntry) {
	l.Messages = append(l.Messages, message)
}

type Session struct {
	Title      string                       `json:"title"`
	Bounds     eskogeom.Rectangle           `json:"bounds"`
	HasBounds  bool                         `json:"hasBounds"`
	Groups     map[string]*Group            `json:"groups"`
	Attributes map[string]map[string]string `json:"attributes"`
}

func (p *Session) OrderedGroups() []*Group {
	var results []*Group
	for _, group := range p.Groups {
		results = append(results, group)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})
	return results
}

func (p *Session) SelectGroup(groupName string) *Group {
	if group, exists := p.Groups[groupName]; exists {
		return group
	}
	group := NewGroup(groupName)
	p.Groups[groupName] = group
	return group
}

func NewGroup(groupName string) *Group {
	defer func() { uniqueGroupIndex++ }()
	return &Group{
		Index:     uniqueGroupIndex,
		Name:      groupName,
		Compounds: make([]*eskogeom.Compound, 0),
	}
}

func (p *Session) updateBounds(point eskogeom.Point) {
	if !p.HasBounds {
		p.Bounds.Origin = point
		p.HasBounds = true
		return
	}
	if point.X < p.Bounds.Origin.X {
		p.Bounds.Width += p.Bounds.Origin.X - point.X
		p.Bounds.Origin.X = point.X
	}
	if point.Y < p.Bounds.Origin.Y {
		p.Bounds.Height += p.Bounds.Origin.Y - point.Y
		p.Bounds.Origin.Y = point.Y
	}
	if point.X > p.Bounds.Origin.X+p.Bounds.Width {
		p.Bounds.Width = point.X - p.Bounds.Origin.X
	}
	if point.Y > p.Bounds.Origin.Y+p.Bounds.Height {
		p.Bounds.Height = point.Y - p.Bounds.Origin.Y
	}
}

func NewSession(title string) *Session {
	return &Session{
		Title:      title,
		Groups:     map[string]*Group{},
		Attributes: make(map[string]map[string]string),
	}
}

func ParseCollection(entries []LogEntry) *Collection {
	doc := Collection{
		Sessions: make([]*Session, 0),
	}
	for _, entry := range entries {
		if entry.Body == "init" {
			if t, ok := entry.Meta["Title"]; ok {
				doc.Sessions = append(doc.Sessions, NewSession(t))
			} else {
				doc.Sessions = append(doc.Sessions, NewSession("Untitled"))
			}
		} else {
			if len(doc.Sessions) == 0 {
				doc.Sessions = append(doc.Sessions, NewSession("Untitled"))
			}
			doc.Sessions[len(doc.Sessions)-1].parse(entry)
		}
	}
	postImportProcessing(&doc)
	return &doc
}

func postImportProcessing(doc *Collection) {
	for _, page := range doc.Sessions {
		for _, group := range page.Groups {
			lastIndex := 0
			for _, compound := range group.Compounds {
				if compound.MetaData["Index"] != "" {
					index, err := strconv.Atoi(compound.MetaData["Index"])
					if err == nil && lastIndex < index {
						lastIndex = index
					}
				}
			}
			group.Partitions = lastIndex + 1
		}
	}
}

func (p *Session) updateBoundsForCompound(compound *eskogeom.Compound) {
	for _, subPath := range compound.SubPaths {
		p.updateBounds(subPath.MoveTo)
		for _, point := range subPath.Points {
			switch v := point.(type) {
			case eskogeom.Point:
				p.updateBounds(v)
			case eskogeom.Quad:
				p.updateBounds(v.C)
				p.updateBounds(v.P)
			case eskogeom.Cubic:
				p.updateBounds(v.C1)
				p.updateBounds(v.C2)
				p.updateBounds(v.P)
			default:
				panic("unknown point type")
			}
		}
	}
}

func (p *Session) parseCompound(entry LogEntry) {
	compound, err := ParseCompound(entry)
	if err != nil {
		log.Fatalln(err)
	}

	groupName, hasGroup := entry.Meta["Layer"]
	if !hasGroup {
		p.SelectGroup(DefaultGroupName).AddCompound(compound)
	} else {
		p.SelectGroup(groupName).AddCompound(compound)
	}

	p.updateBoundsForCompound(compound)
}

func (p *Session) parseText(entry LogEntry) {
	groupName, hasGroup := entry.Meta["Layer"]
	if !hasGroup {
		p.SelectGroup(DefaultGroupName).AddMessage(&entry)
	} else {
		p.SelectGroup(groupName).AddMessage(&entry)
	}
}

func (p *Session) parse(entry LogEntry) {
	if entry.Body[0] == '{' {
		p.parseCompound(entry)
	} else if entry.Body[0:2] == "0x" {
		if _, ok := p.Attributes[entry.Body]; !ok {
			p.Attributes[entry.Body] = make(map[string]string)
		}
		attributes := p.Attributes[entry.Body]

		for key, value := range entry.Meta {
			if vs, ok := attributes[key]; ok {
				// Check if either vs or value contains a comma
				if strings.Contains(vs, ",") || strings.Contains(value, ",") {
					uniqueValues := make(map[string]bool)

					// Add existing values to the set
					for _, v := range strings.Split(vs, ",") {
						uniqueValues[v] = true
					}

					// Add new values to the set, checking for duplicates
					for _, v := range strings.Split(value, ",") {
						uniqueValues[v] = true
					}

					// Combine unique values into a single string
					var combinedValues []string
					for v := range uniqueValues {
						combinedValues = append(combinedValues, v)
					}
					attributes[key] = strings.Join(combinedValues, ",")
				} else if vs != value {
					// If neither vs nor value contains a comma, and they are different, append them
					attributes[key] = vs + "," + value
				}
			} else {
				attributes[key] = value
			}
		}
	} else {
		p.parseText(entry)
	}
}

func ParseCompound(logEntry LogEntry) (*eskogeom.Compound, error) {
	trimmed := strings.TrimSpace(logEntry.Body)

	// Check for opening and closing braces
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return nil, fmt.Errorf("invalid compound string: missing opening or closing brace")
	}

	// Remove the braces and tokenize the string
	trimmed = strings.Trim(trimmed, "{}")
	tokens := strings.Fields(trimmed)

	// Parse the compound's tokens
	compound, err := eskogeom.ParseCompound(tokens)
	if err != nil {
		return nil, err
	}

	// Set the metadata from the log entry
	compound.MetaData = logEntry.Meta

	return compound, nil
}

func ReadCollection(filePath string, opts *ParserOptions) (*Collection, error) {
	data, err := ReadLog(filePath, opts)
	if err != nil {
		return nil, err
	}
	return ParseCollection(data), nil
}
