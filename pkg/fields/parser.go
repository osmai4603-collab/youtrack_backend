package fields

import (
	"strings"
)

// FieldNode يمثل عقدة في شجرة الحقول المطلوبة من معامل fields
type FieldNode struct {
	Name     string
	Children map[string]*FieldNode
}

// HasChildren يتحقق ما إذا كانت العقدة لها حقول فرعية
func (n *FieldNode) HasChildren() bool {
	return len(n.Children) > 0
}

// builtinAliases تعريفات الأسماء المستعارة المدمجة في YouTrack
// (تُستخدم لمن لا يصرّح بها صراحةً داخل معامل fields).
var builtinAliases = map[string]string{
	"permittedUsers":  "id,login,email,fullName,avatarUrl,userType(id,name),name,isEmailVerified,guest,online,banned,banBadge,canReadProfile,isLocked",
	"permittedGroups": "id,name,$type(),auditTargetId,description,allUsersGroup,icon,teamForProject(id,name,icon),isUpdatable,isRemovable",
	"helpdeskContext": "id,name,issuesUrl,pinned,pinnedInHelpdesk,owner(@permittedUsers),$type,query,isUpdatable,shortName",
}

// Parse يحلل معامل fields ويستخرج شجرة الحقول المطلوبة.
// مثال: "id,name,project(id,name,shortName)" => شجرة بـ 3 عقدة
// مثال مع أسماء مستعارة:
//
//	"reporter(@permittedUsers) ; @permittedUsers:id,login,email"
//
// يدعم أيضاً الأسماء المستعارة المدمجة (@permittedUsers, @permittedGroups,
// @helpdeskContext) حتى لو لم تُصرَّح في المعامل نفسه.
// إذا كان fields فارغاً أو غير موجود يُرجع nil (يعني أعد كل شيء)
func Parse(fieldsParam string) *FieldNode {
	if fieldsParam == "" {
		return nil
	}

	mainQuery, declared := splitAliases(fieldsParam)
	if mainQuery == "" {
		return nil
	}

	aliases := make(map[string]string, len(declared)+len(builtinAliases))
	for name, body := range builtinAliases {
		aliases[name] = body
	}
	for name, body := range declared {
		aliases[name] = body
	}

	children := parseFieldList(mainQuery, aliases, make(map[string]bool))
	if len(children) == 0 {
		return nil
	}

	return &FieldNode{
		Name:     "",
		Children: children,
	}
}

// splitAliases يفصل المدخل على ";" ويستخرج تعريفات الأسماء المستعارة
// بصيغة "@name:field,list,..." ويجمع ما تبقى كاستعلام الحقول الرئيسي.
func splitAliases(input string) (string, map[string]string) {
	aliases := make(map[string]string)
	mainParts := make([]string, 0)

	for _, seg := range strings.Split(input, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if name, body, ok := parseAliasDef(seg); ok {
			aliases[name] = body
			continue
		}
		mainParts = append(mainParts, seg)
	}

	return strings.Join(mainParts, ","), aliases
}

// parseAliasDef تحلل قطعة بصيغة "@name:fields" (مع تجاهل أي أقواس كـ "@name(...)")
func parseAliasDef(seg string) (name, body string, ok bool) {
	if !strings.HasPrefix(seg, "@") {
		return "", "", false
	}
	colon := strings.IndexByte(seg, ':')
	if colon < 0 {
		return "", "", false
	}
	namePart := seg[1:colon]
	if i := strings.IndexByte(namePart, '('); i >= 0 {
		namePart = namePart[:i]
	}
	if namePart == "" {
		return "", "", false
	}
	return namePart, seg[colon+1:], true
}

// parseFieldList تحلل قائمة حقول (للمستوى الجذري أو أي قائمة فرعية) وتوسّع
// أي إشارة "@name" من جدول aliases (مع حل التوسعات المتداخلة والحماية من الدورات).
// تُرجع خريطة العقد الفرعية.
func parseFieldList(fieldsParam string, aliases map[string]string, resolving map[string]bool) map[string]*FieldNode {
	children := make(map[string]*FieldNode)

	i := 0
	for i < len(fieldsParam) {
		// تخطي الفاصلة
		if fieldsParam[i] == ',' {
			i++
			continue
		}

		// استخراج اسم الحقل
		start := i
		for i < len(fieldsParam) && fieldsParam[i] != ',' && fieldsParam[i] != '(' && fieldsParam[i] != ')' {
			i++
		}

		fieldName := strings.TrimSpace(fieldsParam[start:i])
		if fieldName == "" {
			continue
		}

		node := &FieldNode{
			Name:     fieldName,
			Children: make(map[string]*FieldNode),
		}

		// إذا وجدنا '(' فهذا حقل فرعي
		if i < len(fieldsParam) && fieldsParam[i] == '(' {
			i++ // تخطي '('
			depth := 1
			subStart := i
			for i < len(fieldsParam) && depth > 0 {
				switch fieldsParam[i] {
				case '(':
					depth++
				case ')':
					depth--
				}
				if depth > 0 {
					i++
				}
			}
			subContent := fieldsParam[subStart:i]
			if i < len(fieldsParam) {
				i++ // تخطي ')'
			}

			node.Children = parseFieldList(subContent, aliases, resolving)
		}

		// اسم مستعار "@name": دمج حقول الـ alias في المستوى الحالي بدل إضافة
		// عقدة باسم حرفي. تُستخدم هذه الآلية مع أنماط مثل
		// reporter(@permittedUsers) أو issueRelatedGroup(@permittedGroups).
		if strings.HasPrefix(fieldName, "@") {
			aliasName := fieldName[1:]
			if body, ok := aliases[aliasName]; ok && !resolving[aliasName] {
				resolving[aliasName] = true
				for k, v := range parseFieldList(body, aliases, resolving) {
					children[k] = v
				}
				delete(resolving, aliasName)
				continue
			}
			// اسم مستعار غير معروف أو ضمن دورة حل → يُحتفظ به كعقدة حرفية
			children[fieldName] = node
			continue
		}

		children[fieldName] = node
	}

	return children
}

// Filter يصفّي البيانات حسب شجرة الحقول المطلوبة
// إذا كان tree فارغاً (nil) يُرجع جميع البيانات كما هي
func Filter(data any, tree *FieldNode) any {
	if tree == nil {
		return data
	}

	switch v := data.(type) {
	case map[string]any:
		return filterMap(v, tree)
	case []any:
		return filterSlice(v, tree)
	default:
		return data
	}
}

// filterMap يصفّي خريطة (JSON object) حسب الحقول المطلوبة.
// المفتاح $type (نوع الكائن) محفوظ دائماً في كل مستوى حتى لو لم يُطلب صراحةً،
// مثل سلوك YouTrack الموثّق في request1.txt.
func filterMap(m map[string]any, tree *FieldNode) map[string]any {
	result := make(map[string]any, len(tree.Children)+1)

	for fieldName, fieldNode := range tree.Children {
		val, exists := m[fieldName]
		if !exists {
			continue
		}

		if fieldNode.HasChildren() {
			// حقل فرعي: تصفية تكراية
			result[fieldName] = Filter(val, fieldNode)
		} else {
			// حقل ورقي: إرجاع القيمة كما هي
			result[fieldName] = val
		}
	}

	// حفظ $type دائماً إن وُجد في الخريطة المصدريّة
	if t, exists := m["$type"]; exists {
		result["$type"] = t
	}

	return result
}

// filterSlice يصفّي مصفوفة (JSON array) - كل عنصر يُصفّي بشكل منفصل
func filterSlice(arr []any, tree *FieldNode) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		result[i] = Filter(item, tree)
	}
	return result
}

// HasField يتحقق ما إذا كان حقل معين موجوداً في شجرة الحقول
func HasField(tree *FieldNode, fieldName string) bool {
	if tree == nil {
		return true // لا يوجد تصفية = الكل موجود
	}
	_, exists := tree.Children[fieldName]
	return exists
}

// GetChild يُرجع العقدة الفرعية لحقل معين
func GetChild(tree *FieldNode, fieldName string) *FieldNode {
	if tree == nil {
		return nil
	}
	return tree.Children[fieldName]
}

// ToJSON يحول شجرة الحقول إلى map[string]any لسهولة الاستخدام في handlers
func ToJSON(tree *FieldNode) map[string]any {
	if tree == nil {
		return nil
	}
	result := make(map[string]any, len(tree.Children))
	for name, child := range tree.Children {
		if child.HasChildren() {
			result[name] = ToJSON(child)
		} else {
			result[name] = true
		}
	}
	return result
}
