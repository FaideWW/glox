## glox - A Lox language interpreter written in Go

Following along with https://craftinginterpreters.com/

Current lox grammar:

```
program        → declaration* EOF ;

declaration    → varDecl | statement ;

varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;

statement      → exprStmt | printStmt | block;

exprStmt       → expression ";" ;
printStmt      → "print" expression ";" ;
block          → "{" declaration* "}" ;

expression     → assignment ( ( "," ) assignment )* ;
assignment     → IDENTIFIER "=" assignment
               | condition ;
condition      → equality ( ( "?" ) condition ( ":" ) condition )? ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary
               | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | "(" expression ")"
               | IDENTIFIER ;
```
