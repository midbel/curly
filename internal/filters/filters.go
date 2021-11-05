package filters

func First() (interface{}, error) {
  return FirstN()
}

func Last() (interface{}, error) {
  return LastN()
}

func FirstN() (interface{}, error) {
  return nil, nil
}

func LastN() (interface{}, error) {
  return nil, nil
}

func Reverse() (interface{}, error) {
  return nil, nil
}
