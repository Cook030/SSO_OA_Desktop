package shared

import (
	"errors"
	"regexp"
)

// ValidateEmployeeName 校验员工姓名长度
func ValidateEmployeeName(name string) error {
	if len(name) < 2 || len(name) > 64 {
		return errors.New("员工姓名长度需在2-64字符之间")
	}
	return nil
}

// ValidateAccount 校验员工账号(姓名全拼)
func ValidateAccount(account string) error {
	if len(account) < 2 || len(account) > 64 {
		return errors.New("员工账号长度需在2-64字符之间")
	}
	return nil
}

// ValidatePhone 校验中国大陆手机号，为空则跳过校验
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil
	}
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	if !match {
		return errors.New("手机号格式不正确")
	}
	return nil
}

// ValidateEmailPrefix 校验邮箱前缀，为空则跳过校验
func ValidateEmailPrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	match, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, prefix)
	if !match {
		return errors.New("邮箱前缀只能包含字母、数字、点、下划线和短横线")
	}
	return nil
}

// ValidateDepartment 校验部门
func ValidateDepartment(dept string) error {
	if dept == "" {
		return errors.New("所属部门不能为空")
	}
	return nil
}

// ValidatePassword 校验密码长度
func ValidatePassword(password string) error {
	if len(password) < 6 || len(password) > 64 {
		return errors.New("密码长度需在6-64字符之间")
	}
	return nil
}

// ValidatePlatformName 校验平台名称长度
func ValidatePlatformName(name string) error {
	if len(name) < 2 || len(name) > 128 {
		return errors.New("平台名称长度需在2-128字符之间")
	}
	return nil
}

// ValidatePlatformLink 校验平台链接长度
func ValidatePlatformLink(link string) error {
	if len(link) < 2 || len(link) > 128 {
		return errors.New("平台链接长度需在2-128字符之间")
	}
	return nil
}
