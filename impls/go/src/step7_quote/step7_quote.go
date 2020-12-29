package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/jamesroutley/mal/impls/go/src/core"
	"github.com/jamesroutley/mal/impls/go/src/environment"
	"github.com/jamesroutley/mal/impls/go/src/printer"
	"github.com/jamesroutley/mal/impls/go/src/reader"
	"github.com/jamesroutley/mal/impls/go/src/types"
)

var debugExpressions = []string{
	// "(= (quote abc) (quote abc))",
}

func main() {
	env := environment.NewEnv()
	for _, item := range core.Namespace {
		env.Set(item.Symbol.Value, item.Func)
	}

	// Builtin functions defined in lisp
	_, err := Rep("(def! not (fn* (a) (if a false true)))", env)
	if err != nil {
		log.Fatal(err)
	}

	_, err = Rep("(def! load-file (fn* (f) (eval (read-string (+ \"(do \" (slurp f) \"\nnil)\")))))", env)
	if err != nil {
		log.Fatal(err)
	}

	if len(debugExpressions) != 0 {
		for _, expr := range debugExpressions {
			fmt.Printf("user> %s\n", expr)
			output, err := Rep(expr, env)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(output)
		}
		return
	}

	_, err = Rep(`(def! lst (quote (b c)))`, env)
	if err != nil {
		log.Fatal(err)
	}
	_, err = Rep(`(quasiquoteexpand (a (unquote lst) d))`, env)
	if err != nil {
		log.Fatal(err)
	}

	// Eval function. Needs to be here, because it closes over `env`
	env.Set("eval", &types.MalFunction{
		Func: func(args ...types.MalType) (types.MalType, error) {
			return Eval(args[0], env)
		},
	})

	// code := `"abc"`
	// ast, err := Read(code)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// reader.DebugType(ast)
	// fmt.Println(Eval(ast, env))

	// return
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "user> ",
		HistoryFile: "/Users/jamesroutley/.malhistory",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		// TODO: investigate config of this. For example, it would be nice to
		// store a history
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		line = strings.TrimSuffix(line, "\n")
		output, err := Rep(line, env)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(output)
	}
}

// Read tokenizes and parses source code
func Read(s string) (types.MalType, error) {
	return reader.ReadStr(s)
}

// Eval evaulates a piece of parsed code.
// The way code is evaluated depends on its structure.
//
// 1. Special forms - these are language-level features, which behave
// differently to normal functions. These include `def!` and `let*`. For
// example, certain elements in the argument list might be evaluated
// differently (or not at all)
// 2. Symbols: evaluated to their corresponding value in the environment `env`
// 3. Lists: by default, they're treated as function calls - each item is
// evaluated, and the first item (the function itself) is called with the rest
// of the items as arguments.
func Eval(ast types.MalType, env *environment.Env) (types.MalType, error) {
top:
	// First - check if ast is a list. If it isn't we can evaluate it as an
	// atom and return
	list, ok := ast.(*types.MalList)
	if !ok {
		return evalAST(ast, env)
	}
	if len(list.Items) == 0 {
		return ast, nil
	}

	// Ok, AST is a list. Lists can contain function calls, macros, special
	// forms. Here we handle those cases.

	// First, macros. A macro modifies Lisp source code, so we need to expand
	// them before we continue evaluating.
	{
		expandedAST, err := macroExpand(ast, env)
		if err != nil {
			return nil, err
		}

		// Check if the ast is still a list after the macro expansion. If it
		// isn't, we just return evalAST, like we did for non-lists above.
		// If it is, continue.
		switch expandedAST.(type) {
		case *types.MalList:
			ast = expandedAST
			// continue
		default:
			return evalAST(expandedAST, env)
		}
	}

	// Some special forms are tail call optimised. Instead of recusively
	// calling Eval, they return a new `ast` and `env`, and we loop back to
	// the top of this function.
	// TODO: I think we can pass list here, rather than ast
	if operator, args, ok := isTCOSpecialForm(ast); ok {
		newAST, newEnv, err := evalTCOSpecialForm(operator, args, env)
		if err != nil {
			return nil, err
		}
		ast = newAST
		env = newEnv
		// XXX: The other option here is to wrap this function body if a while
		// loop, and `continue` here. They've equivalent because all other
		// branches return. Using a goto seems somewhat nicer though??
		goto top
	}

	if operator, args, ok := isSpecialForm(ast); ok {
		return evalSpecialForm(operator, args, env)
	}

	// Apply phase - evaluate all elements in the list, then call the first
	// as a function, with the rest as arguments
	evaluated, err := evalAST(list, env)
	if err != nil {
		return nil, err
	}

	evaluatedList, ok := evaluated.(*types.MalList)
	if !ok {
		return nil, fmt.Errorf("list did not evaluate to a list")
	}

	function, ok := evaluatedList.Items[0].(*types.MalFunction)
	if !ok {
		return nil, fmt.Errorf("first item in list isn't a function")
	}

	if !function.TailCallOptimised {
		return function.Func(evaluatedList.Items[1:]...)
	}

	// Function is tail call optimised.
	// Construct the correct environment it should be run in
	childEnv := environment.NewChildEnv(
		function.Env.(*environment.Env), function.Params, evaluatedList.Items[1:],
	)

	ast = function.AST
	env = childEnv
	goto top
}

// Print prints the AST as a human readable string. It's not inteded for debugging
func Print(s types.MalType) string {
	return printer.PrStr(s)
}

// Rep - read, evaluate, print
func Rep(s string, env *environment.Env) (string, error) {
	t, err := Read(s)
	if err != nil {
		return "", err
	}
	t, err = Eval(t, env)
	if err != nil {
		return "", err
	}
	s = Print(t)
	return s, nil
}

// evalAST implements the evaluation rules for normal expressions. Any special
// cases are handed above us, in the Eval function. This function is an
// implementation detail of Eval, and shoulnd't be called apart from by it.
func evalAST(ast types.MalType, env *environment.Env) (types.MalType, error) {
	switch tok := ast.(type) {
	case *types.MalSymbol:
		value, err := env.Get(tok.Value)
		if err != nil {
			return nil, err
		}
		return value, nil
	case *types.MalList:
		items := make([]types.MalType, len(tok.Items))
		for i, item := range tok.Items {
			evaluated, err := Eval(item, env)
			if err != nil {
				return nil, err
			}
			items[i] = evaluated
		}
		return &types.MalList{
			Items: items,
		}, nil
	}
	return ast, nil
}

func isTCOSpecialForm(ast types.MalType) (operator *types.MalSymbol, args []types.MalType, ok bool) {
	tok, ok := ast.(*types.MalList)
	if !ok {
		return nil, nil, false
	}
	items := tok.Items
	if len(items) == 0 {
		return nil, nil, false
	}

	operator, ok = items[0].(*types.MalSymbol)
	if !ok {
		return nil, nil, false
	}

	switch operator.Value {
	case "let*", "if", "do", "quasiquote":
		return operator, items[1:], true
	}

	return nil, nil, false
}

func isSpecialForm(ast types.MalType) (operator *types.MalSymbol, args []types.MalType, ok bool) {
	tok, ok := ast.(*types.MalList)
	if !ok {
		return nil, nil, false
	}
	items := tok.Items
	if len(items) == 0 {
		return nil, nil, false
	}

	operator, ok = items[0].(*types.MalSymbol)
	if !ok {
		return nil, nil, false
	}

	switch operator.Value {
	case "fn*", "def!", "quote", "quasiquoteexpand", "defmacro!", "macroexpand":
		return operator, items[1:], true
	}

	return nil, nil, false
}

// Some special forms end in an evaluation. We could implement this by
// recusively calling `Eval` (it's recusive because evalTCOSpecialForm is
// called by Eval), but that can lead to stack overflow issues. Instead, we
// tail call optimise by returning a new AST to evaluate, and a new environment
// to invaluate it in. Eval loops back to the beginning of the function and
// re-runs itself using these new params.
func evalTCOSpecialForm(
	operator *types.MalSymbol, args []types.MalType, env *environment.Env,
) (newAST types.MalType, newEnv *environment.Env, err error) {
	switch operator.Value {

	// Creates a new environment with certain variables set, then evaluates a
	// statement in that environment.
	// e.g:
	//
	// > (let* (a 1 b (+ a 1)) b)
	// 2 ; a == b, b == a+1 == 2
	case "let*":
		if len(args) != 2 {
			return nil, nil, fmt.Errorf("let* takes 2 args")
		}
		bindingList, ok := args[0].(*types.MalList)
		if !ok {
			return nil, nil, fmt.Errorf("let*: first arg isn't a list")
		}
		if len(bindingList.Items)%2 != 0 {
			return nil, nil, fmt.Errorf("let*: first arg doesn't have an even number of items")
		}

		childEnv := env.ChildEnv()
		for i := 0; i < len(bindingList.Items); i += 2 {
			key, ok := bindingList.Items[i].(*types.MalSymbol)
			if !ok {
				return nil, nil, fmt.Errorf("let*: binding list: arg %d isn't a symbol", i)
			}
			value, err := Eval(bindingList.Items[i+1], childEnv)
			if err != nil {
				return nil, nil, err
			}
			childEnv.Set(key.Value, value)
		}

		// Finally, return the last arg as the new AST to be evaluated, and the
		// newly constructed env as the environment
		return args[1], childEnv, nil

	// Evaluates the elements in the arg list and returns the final result.
	// For TCO, we eval all but the last argument here, then return the last
	// argument to be evaluated in the main Eval loop.
	case "do":
		for _, arg := range args[:len(args)-1] {
			var err error
			_, err = Eval(arg, env)
			if err != nil {
				return nil, nil, err
			}
		}
		return args[len(args)-1], env, nil

	// Evaluate first param. If not `nil` or `false`, return the second param
	// to be evaluated. If it is, return the third param to be evaluated, or
	// `nil` if none is supplied. If none is supplied, the `nil` value is
	// evalulated, but just evaluates to `nil`.
	case "if":
		if numArgs := len(args); numArgs != 2 && numArgs != 3 {
			return nil, nil, fmt.Errorf("if statements must have two or three arguments, got %d", numArgs)
		}
		condition, err := Eval(args[0], env)
		if err != nil {
			return nil, nil, err
		}
		if IsTruthy(condition) {
			return args[1], env, nil
		}

		if len(args) == 3 {
			return args[2], env, nil
		}

		return &types.MalNil{}, env, nil

	case "quasiquote":
		ast, err := quasiquote(args[0])
		if err != nil {
			return nil, nil, err
		}
		return ast, env, nil

	default:
		return nil, nil, fmt.Errorf("unexpected tail call optimised special form: %s", operator.Value)
	}
}

func evalSpecialForm(
	operator *types.MalSymbol, args []types.MalType, env *environment.Env,
) (types.MalType, error) {
	switch operator.Value {

	// Assigns a value to a symbol in the current environment
	// e.g:
	//
	// > (def a 10)
	// 10
	// > a
	// 10
	case "def!":
		if len(args) != 2 {
			return nil, fmt.Errorf("def! takes 2 args")
		}
		key, ok := args[0].(*types.MalSymbol)
		if !ok {
			return nil, fmt.Errorf("def!: first arg isn't a symbol")
		}
		value, err := Eval(args[1], env)
		if err != nil {
			return nil, err
		}
		env.Set(key.Value, value)
		return value, nil

	// Create a new function.
	//
	// e.g:
	// > (def! add1 (fn* (a) (+ a 1)))
	// #<function>
	// > (add1 2)
	// 3
	case "fn*":
		if len(args) != 2 {
			return nil, fmt.Errorf("fn* statements must have two arguments, got %d", len(args))
		}

		// arguments is the first argument supplied to the fn* function (e.g.
		// `(a)` in the example above)
		arguments, ok := args[0].(*types.MalList)
		if !ok {
			return nil, fmt.Errorf("fn* statements must have a list as the first arg")
		}
		// Cast it from a list of MalType to a list of MalSymbol
		binds := make([]*types.MalSymbol, len(arguments.Items))
		for i, a := range arguments.Items {
			bind, ok := a.(*types.MalSymbol)
			if !ok {
				// TODO: improve this - say which argument isn't a symbol
				return nil, fmt.Errorf("fn* statements must have a list of symbols as the first arg")
			}
			binds[i] = bind
		}

		// TODO: recomment this
		return &types.MalFunction{
			TailCallOptimised: true,
			AST:               args[1],
			Params:            binds,
			Env:               env,
			// This Go function is what's run when the Lisp function is
			// run. When the Lisp function is run, we create a new environment,
			// which binds the Lisp function's arguments to the parameters
			// defined when the function was defined.
			Func: func(exprs ...types.MalType) (types.MalType, error) {
				childEnv := environment.NewChildEnv(
					env, binds, exprs,
				)
				return Eval(args[1], childEnv)
			},
		}, nil

	case "quote":
		return args[0], nil

	case "quasiquoteexpand":
		return quasiquote(args[0])

	// Creates a new macro
	case "defmacro!":
		if len(args) != 2 {
			return nil, fmt.Errorf("defmacro! takes 2 args")
		}
		key, ok := args[0].(*types.MalSymbol)
		if !ok {
			return nil, fmt.Errorf("defmacro!: first arg isn't a symbol")
		}
		value, err := Eval(args[1], env)
		if err != nil {
			return nil, err
		}
		function, ok := value.(*types.MalFunction)
		if !ok {
			return nil, fmt.Errorf("defmacro!: second arg isn't a function definition")
		}
		function.IsMacro = true
		env.Set(key.Value, function)
		return function, nil

	// Macroexpand expands a macro and returns the expanded form. Useful for
	// debugging macros
	case "macroexpand":
		return macroExpand(args[0], env)

	// XXX: if you add a case here, you also need to add it to `isSpecialForm`

	default:
		return nil, fmt.Errorf("unexpected special form: %s", operator.Value)
	}
}

// IsTruthy returns a type's truthiness. Currently: it's falsy if the type is
// `nil` or the boolean 'false'. All other values are truthy.
func IsTruthy(t types.MalType) bool {
	switch token := t.(type) {
	case *types.MalNil:
		return false
	case *types.MalBoolean:
		return token.Value
	}
	return true

}

func quasiquote(ast types.MalType) (types.MalType, error) {
	list, ok := ast.(*types.MalList)
	if !ok {
		// `ast` isn't a list, which means it can't be an unquoted form. Return
		// its quoted form. Here, we quote it regardless of its type.
		// Quoting forms such as ints and strings is redundant - quoting
		// prevents the evaluator from evaluating a form, but those forms
		// evaluate to themselves, so there's no difference caused by quoting
		// them. However, there's also no harm harm in doing so.
		// This return statements returns the AST version of (quote <ast>)
		return &types.MalList{
			Items: []types.MalType{
				&types.MalSymbol{Value: "quote"},
				ast,
			},
		}, nil
	}

	// Okay - ast is a list
	items := list.Items

	// If the list has no items, return it unmodified
	if len(items) == 0 {
		return ast, nil
	}

	// If the first item in the list is the function `unquote`, return the
	// first argument without quoting it.
	if symbol, ok := items[0].(*types.MalSymbol); ok && symbol.Value == "unquote" {
		return list.Items[1], nil
	}

	// Okay - ast is a list, than hasn't been unquoted
	quasiquoted := &types.MalList{}

	for i := len(items) - 1; i >= 0; i-- {
		element := items[i]

		// TODO: implement `splice-unquote` functionality
		if args, ok := isSpliceUnquoteForm(element); ok {
			quasiquoted = &types.MalList{
				Items: []types.MalType{
					&types.MalSymbol{Value: "concat"},
					args[0],
					quasiquoted,
				},
			}
			continue
		}

		quasiqutoedElement, err := quasiquote(element)
		if err != nil {
			return nil, err
		}

		quasiquoted = &types.MalList{
			Items: []types.MalType{
				&types.MalSymbol{Value: "cons"},
				quasiqutoedElement,
				quasiquoted,
			},
		}

	}
	return quasiquoted, nil
}

func isSpliceUnquoteForm(ast types.MalType) (spliceUnquoteArgs []types.MalType, ok bool) {
	list, ok := ast.(*types.MalList)
	if !ok {
		return nil, false
	}
	items := list.Items
	if len(items) == 0 {
		return nil, false
	}
	symbol, ok := items[0].(*types.MalSymbol)
	if !ok {
		return nil, false
	}
	if symbol.Value == "splice-unquote" {
		return items[1:], true
	}
	return nil, false
}

func isMacroCall(ast types.MalType, env *environment.Env) bool {
	list, ok := ast.(*types.MalList)
	if !ok {
		return false
	}
	items := list.Items
	if len(items) == 0 {
		return false
	}
	symbol, ok := items[0].(*types.MalSymbol)
	if !ok {
		return false
	}
	value, err := env.Get(symbol.Value)
	if err != nil {
		// This looks dangerous, but is okay - the only error this function
		// returns is a not found when the symbol isn't defined in any
		// environment
		return false
	}
	function, ok := value.(*types.MalFunction)
	if !ok {
		return false
	}
	return function.IsMacro
}

func macroExpand(ast types.MalType, env *environment.Env) (types.MalType, error) {
	for isMacroCall(ast, env) {
		// TODO: isMacroCall could return the macro function, which would save
		// the casting below
		//
		// Don't check the ok value here because we've validated that ast is a
		// list in isMacroCall. If it's not, something very strange has
		// happened
		list := ast.(*types.MalList)
		// Again, we've already checked this - skip ok checking
		macroName := list.Items[0].(*types.MalSymbol)

		macroNameValue, err := env.Get(macroName.Value)
		if err != nil {
			// Shouldn't happen - we've already validated this above
			return nil, err
		}
		macroFunc := macroNameValue.(*types.MalFunction)

		newAst, err := macroFunc.Func(list.Items[1:]...)
		if err != nil {
			return nil, err
		}

		// Set the evaulated macro to `ast` and loop back - this lets us
		// iteratively expand nested macros
		ast = newAst
	}
	return ast, nil
}
