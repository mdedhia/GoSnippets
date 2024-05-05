package gosnippets

/*
You are climbing a staircase. It takes n steps to reach the top.

Each time you can either climb 1 or 2 steps. In how many distinct ways can you climb to the top?
 

Example 1:

Input: n = 2
Output: 2
Explanation: There are two ways to climb to the top.
1. 1 step + 1 step
2. 2 steps

Example 2:

Input: n = 3
Output: 3
Explanation: There are three ways to climb to the top.
1. 1 step + 1 step + 1 step
2. 1 step + 2 steps
3. 2 steps + 1 step 

Constraints:

1 <= n <= 45
*/

func climb(n int, memo map[int]int) int {
    ways, ok := memo[n]
    if ok {
        return ways
    }

    if n == 0 {
        return 1
    }
    if n < 0 {
        return 0
    } 

    memo[n] = climb(n - 2, memo) + climb(n - 1, memo)
    return memo[n]
}

func climbStairs(n int) int {
	if n < 1 || n > 45 {
		// fmt.Println("Invalid input. Expecting: 1 <= n <= 45")	
		return -1
	}

    memo := make(map[int]int)	// Memoize the solution for faster performance
    
    return climb(n, memo)
}
