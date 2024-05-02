package gosnippets

/*
Given a linked list, swap every two adjacent nodes and return its head. 
You must solve the problem without modifying the values in the list's nodes 
(i.e., only nodes themselves may be changed.)

Example 1:

Input: head = [1,2,3,4]
Output: [2,1,4,3]

Example 2:

Input: head = []
Output: []
Example 3:

Input: head = [1]
Output: [1] 

Constraints:

The number of nodes in the list is in the range [0, 100].
0 <= Node.val <= 100
*/

//Definition for singly-linked list.
type ListNode struct {
    Val int
    Next *ListNode
}

func swapPairs(head *ListNode) *ListNode {
    if head == nil {
        return nil
	} else if head.Next == nil {
		return head
	}
        
	n1, n2, startPtr := head, head.Next, head
	head = head.Next                // Initialize head to second node
	for n1 != nil && n2 != nil {
		startPtr.Next = n2
		n1.Next = n2.Next
		n2.Next = n1
		startPtr = n1

		n1 = n1.Next
		if n1 != nil {
			n2 = n1.Next
		} else {
			break
		}
	}
    return head
}