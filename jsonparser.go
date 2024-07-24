package main

import (
	"fmt"
	"os"
	"io"
	"strings"
	"errors"
	"unicode"
	//"reflect"
)

type JSONObject map[string]any
type TokenType int 

const (
	TokenString = iota
	TokenFloat
	TokenArr
	TokenComma
	TokenColon
)

type Token struct {
	Type TokenType
	Value string
	Expect rune
	Length int
}

type Profile struct {
	Name string
	Email string
	Address string
	Favorite_Numbers []float64
	Age float64
	Avatar string
}

func main() {
	runParser("json_examples.txt", "examples_output.txt")

	//runFormatted()
}

/*
// NOTE: Unfinished code, attempting to set up a structure to capture
// expected JSON objects. Have not yet had time to finish.
func runFormatted() {
	var profiles []Profile

	file_input := "test_examples.txt"
	file_output := "formatted_output.txt"

	// open file
	json_file, err := os.Open(file_input)
	if err != nil {
		fmt.Println("opening file error", err)
	}
	
	
	// convert all file lines into a single string
	// could probably be optimized
	json_string, err := io.ReadAll(json_file)		
	if err != nil {
		fmt.Println("reading file error", err)
	}

	json_file.Close()
	obj_list, err := parseJSON(string(json_string))
	
	if err != nil {
		fmt.Println("error in parsing", err)
	}
	
	f, err := os.Create(file_output)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, m := range obj_list {
		var profile Profile

		// m is a map[string]interface.
		// loop over keys and values in the map.
		for k, v := range m {
			prof := reflect.ValueOf(profile)
			field := prof.FieldByName(k)

			value_type := reflect.ValueOf(v)
			
			field.Set(value_type)

		
			}


			fmt.Println(profile.Name)
			profiles = append(profiles, profile)
		}



		f.Close()
	}

*/


func runParser(file_input string, file_output string) {

	// open file
	json_file, err := os.Open(file_input)
	if err != nil {
		fmt.Println("opening file error", err)
	}


	// convert all file lines into a single string
	// could probably be optimized
	json_string, err := io.ReadAll(json_file)
	if err != nil {
		fmt.Println("reading file error", err)
	}

	json_file.Close()
	obj_list, err := parseJSON(string(json_string))

	if err != nil {
		fmt.Println("error in parsing", err)
	}

	f, err := os.Create(file_output)
	if err != nil {
		fmt.Println(err)
		return
	}


	// loop over elements of slice
	for i, m := range obj_list {
		if i != 0{
			_, err := fmt.Fprintf(f, "\n")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		_, err := fmt.Fprintf(f, "Object %d\n", i + 1)
		if err != nil {
			fmt.Println(err)
			return
		}

		// m is a map[string]interface.
		// loop over keys and values in the map.
		for k, v := range m {
			
			/*
			if v == "" {
				v = "None"
			}
				*/
			
			_, err := fmt.Fprintf(f, "%s: %v\n", k, v)
			if err != nil {
				fmt.Println(err)
				return
			}

		}

		
	}
	f.Close()

}

func parseJSON(json_string string) ([]JSONObject, error) {

	var obj_list []JSONObject

	// trim leading/trailing white space
	json_string = strings.TrimSpace(json_string)

	// check what we're working with

	if strings.HasPrefix(json_string, "[") && strings.HasSuffix(json_string, "]"){
		// array case
		// trim off []
		json_string := json_string[1:len(json_string) - 1]

		json_runes := []rune(json_string)

		// find all objects, ensure that they are in good form
		// then parse each object
		obj_start := 0
		new_obj_allowed := true

		for i := 0; i < len(json_runes) ; i++{
			
			// whenever we're looking for new object, i and obj_start will be equal
			// last_rune SHOULD also be equal to ',' between objects
			if i == obj_start {
				// manually remove preceding whitespace if looking for new object
				if unicode.IsSpace(json_runes[i]) {
					obj_start++
					continue
				// found a comma between objects -- allow new object to be found
				} else if json_runes[i] == ',' && !new_obj_allowed{
					new_obj_allowed = true
					obj_start++
					continue
				// case in which we've found beginning of new object
				} else if json_runes[i] == '{' && new_obj_allowed {
					// do not increment object starting
					new_obj_allowed = false
					continue
				} else {
					// only commas (sometimes) and whitespace may exist between objects
					// if any other character is found, string is improperly formatted
					return obj_list, errors.New("Improperly formatted JSON")
				}
			// if already parsing an object

			} else if json_runes[i] == '{' {
				// nested object - not allowed
				return obj_list, errors.New("Improperly formatted JSON")
			} else if json_runes[i] == '}'{
				// found end of object yay! :) 
				obj, err := parseObject(string(json_runes[obj_start:i + 1]))

				if err != nil {
					return obj_list, err
				}
				
				obj_list = append(obj_list, obj)
				
				// make next object begin at next rune
				// this should either be a comma or whitespace
				obj_start = i + 1
			} 
			
		}

	} else if strings.HasPrefix(json_string, "{") && strings.HasSuffix(json_string, "}") {
		// single object case

		obj, err := parseObject(json_string)
		
		if err != nil {
			return obj_list, err
		}

		obj_list = append(obj_list, obj)
		

	} else {

		return obj_list, errors.New("Improperly formatted JSON")

	}

	return obj_list, nil
}

func parseObject(json_object string) (JSONObject, error) {

	object := make(JSONObject)

	// trim brackets and whitespace 
	json_object = strings.TrimSpace(json_object[1:len(json_object) - 1])



	// we are expecting a format of "string" : string/int/array (,)

	i := 0
	for i < len(json_object) {

		property, err := findNextToken(json_object[i:])

		if err != nil {
			
			fmt.Println(err)
			return object, err
		}

		colon, err := findNextToken(json_object[i + property.Length:])

		
		if err != nil{
			
			fmt.Println(err)
			return object, err
		} else if colon.Type != TokenColon {
			
			return object, errors.New("Improperly formatted JSON")
		} 

		value, err := findNextToken(json_object[i + property.Length + 1:])

		if err != nil {
			
			fmt.Println(err)
			return object, err
		}

		i = i + property.Length + value.Length + 1

		object[property.Value] = value.Value


		// note that we can expect i == len(json_object) - 1 at this point if 
		// we're at end of file and it's formatted correctly

		if i != len(json_object) {

			comma, err := findNextToken(json_object[i:])

			if err != nil{
				
				fmt.Println(err)
				return object, err
			} else if comma.Type != TokenComma {
				fmt.Println("Here?")
				return object, errors.New("Improperly formatted JSON")
			}
		}
		i++


		
	}



	// we should now have just key/value pairs

	// fmt.Println(json_object)

	return object, nil


}

func findNextToken(json_object string) (Token, error){
	// given a string, find the next token in it
	// if no tokens found, return error

	json_runes := []rune(json_object)

	whitespace := 0
	// first character determines what token we're looking for
	flag := json_runes[whitespace]


	for whitespace < len(json_runes){
		flag = json_runes[whitespace]
		if !unicode.IsSpace(flag){
			break
		}
		whitespace++
		
	
	}

	var token Token


	if flag == ','{
		token = Token{Type: TokenComma, Length: 1}
	} else if flag == ':' {
		token = Token{Type: TokenColon, Length: 1}
	} else if flag == '[' {
		token = Token{Type: TokenArr, Expect: ']'}
	} else if flag == '"' {
		token = Token{Type: TokenString, Expect: '"'}
	} else if unicode.IsDigit(flag) {
		token = Token{Type: TokenFloat, Expect: ','}
	} else {
		return token, errors.New("Improperly formatted JSON")
	}


	for i := whitespace + 1; i < len(json_runes); i++ {
		if json_runes[i] == token.Expect {
			if token.Type == TokenFloat {
				token.Value = string(json_runes[whitespace:i])
				token.Length = i
			} else{
				token.Value = string(json_runes[whitespace + 1:i])
				
				token.Length = i  + 1
			}
			return token, nil
		}

	}

	if token.Type == TokenFloat {
		token.Value = json_object
		return token, nil
	}

	return token, nil

}

