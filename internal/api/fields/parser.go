package fields

import (
	"net/url"
	"strings"
)

// FieldTree يمثّل شجرة الحقول المستخرجة من معامل fields.
type FieldTree struct {
	Name     string
	Children map[string]*FieldTree
}

// NewFieldTree ينشئ عقدة شجرة حقول جديدة.
func NewFieldTree(name string) *FieldTree {
	return &FieldTree{
		Name:     name,
		Children: make(map[string]*FieldTree),
	}
}

// Has يتحقق مما إذا كان المسار المحدد موجوداً في شجرة الحقول.
// يدعم المسارات المدمجة بنقاط، مثلاً: "profiles.general.timezone" أو "profiles"
func (t *FieldTree) Has(path string) bool {
	if t == nil {
		return false
	}
	if path == "" {
		return true
	}
	parts := strings.Split(path, ".")
	curr := t
	for _, part := range parts {
		if curr.Children == nil {
			return false
		}
		next, exists := curr.Children[part]
		if !exists {
			return false
		}
		curr = next
	}
	return true
}

// Child يعيد شجرة الحقول الفرعية لاسم حقل معين.
func (t *FieldTree) Child(name string) *FieldTree {
	if t == nil || t.Children == nil {
		return nil
	}
	return t.Children[name]
}

// IsEmpty يتحقق مما إذا كانت شجرة الحقول فارغة.
func (t *FieldTree) IsEmpty() bool {
	return t == nil || len(t.Children) == 0
}

// Parse يحلل السلسلة النصية لمعامل fields ويبني الـ FieldTree مع توسيع القوالب @templates.
func Parse(rawFields string) *FieldTree {
	if rawFields == "" {
		return nil
	}

	// 1. فك تشفير الـ URL Encoding
	decoded, err := url.QueryUnescape(rawFields)
	if err != nil {
		decoded = rawFields
	}

	// 2. تقسيم السلسلة إلى الاستعلام الأساسي والقوالب @templates
	parts := splitTopLevel(decoded, ';')
	if len(parts) == 0 {
		return nil
	}

	mainExpr := strings.TrimSpace(parts[0])
	templates := make(map[string]string)

	for i := 1; i < len(parts); i++ {
		p := strings.TrimSpace(parts[i])
		if strings.HasPrefix(p, "@") {
			colonIdx := strings.Index(p, ":")
			if colonIdx > 0 {
				tplName := strings.TrimSpace(p[:colonIdx])
				tplBody := strings.TrimSpace(p[colonIdx+1:])
				templates[tplName] = tplBody
			}
		}
	}

	// 3. بناء شجرة الحقول وتوسيع القوالب
	root := NewFieldTree("")
	parseExpression(mainExpr, root, templates, 0)
	return root
}

// parseExpression يحلل قائمة الحقول المفصولة بفواصل مع مراعاة الأقواس والقوالب.
func parseExpression(expr string, parent *FieldTree, templates map[string]string, depth int) {
	if depth > 20 || expr == "" { // منع الحلقات التكرارية اللانهائية
		return
	}

	items := splitTopLevel(expr, ',')
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// إزالة معلومات التقسيم مثل usersInTeam:0:16
		colonIdx := strings.Index(item, ":")
		parenIdx := strings.Index(item, "(")

		var fieldName string
		var subExpr string

		if parenIdx != -1 && (colonIdx == -1 || parenIdx < colonIdx) {
			fieldName = strings.TrimSpace(item[:parenIdx])
			endParenIdx := findMatchingParen(item, parenIdx)
			if endParenIdx > parenIdx {
				subExpr = item[parenIdx+1 : endParenIdx]
			}
		} else if colonIdx != -1 && (parenIdx == -1 || colonIdx < parenIdx) {
			// e.g. usersInTeam:0:16(@users)
			fieldName = strings.TrimSpace(item[:colonIdx])
			if parenIdx != -1 {
				endParenIdx := findMatchingParen(item, parenIdx)
				if endParenIdx > parenIdx {
					subExpr = item[parenIdx+1 : endParenIdx]
				}
			}
		} else {
			fieldName = item
		}

		// تنظيف اسم الحقل من أي زوائد
		fieldName = strings.TrimSuffix(fieldName, "()")
		if fieldName == "" {
			continue
		}

		// فحص إذا كان الحقل أو تعبيره الفرعي قالبًا @template
		if strings.HasPrefix(fieldName, "@") {
			if tplContent, exists := templates[fieldName]; exists {
				parseExpression(tplContent, parent, templates, depth+1)
			}
			continue
		}

		node, exists := parent.Children[fieldName]
		if !exists {
			node = NewFieldTree(fieldName)
			parent.Children[fieldName] = node
		}

		if subExpr != "" {
			subTrimmed := strings.TrimSpace(subExpr)
			if strings.HasPrefix(subTrimmed, "@") {
				if tplContent, exists := templates[subTrimmed]; exists {
					parseExpression(tplContent, node, templates, depth+1)
				}
			} else {
				parseExpression(subExpr, node, templates, depth+1)
			}
		}
	}
}

// splitTopLevel يقسم النص بفاصل محدد مع تجاهل الفواصل داخل الأقواس ().
func splitTopLevel(s string, delimiter rune) []string {
	var results []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		if r == '(' {
			depth++
			current.WriteRune(r)
		} else if r == ')' {
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		} else if r == delimiter && depth == 0 {
			results = append(results, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		results = append(results, current.String())
	}
	return results
}

// findMatchingParen يجد قوس الإغلاق المقابل لقوس الفتح عند startIdx.
func findMatchingParen(s string, startIdx int) int {
	depth := 0
	for i := startIdx; i < len(s); i++ {
		if s[i] == '(' {
			depth++
		} else if s[i] == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return len(s) - 1
}
