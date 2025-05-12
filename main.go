package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"

	"paxos-simulator/node"
)

func print_and_scan(p string, s any) {
	fmt.Printf("> %v", p)
	fmt.Scanln(s)
}

func main() {
	num_replicas := 1
	print_and_scan("Enter number of replicas...: ", &num_replicas)
	if num_replicas % 2 == 0 {
		fmt.Println("need even number of replicas")
		os.Exit(-1)
	}
	nodes := make([]*node.Node, num_replicas)
	for i := 0; i < len(nodes); i++ {
		nodes[i] = &node.Node {
			NumReplicas: num_replicas,
		} 
		nodes[i].Start()
	}
	fmt.Printf("initialized %v nodes, N_0...N_%v\n", num_replicas, num_replicas - 1)
	fmt.Println("Try to reach consensus.") 
	for {
		print_state(nodes)
		fmt.Println("Choose a node and an action. Possible actions include:\nA_0: construct prepare message (first step to have a node act as a proposer)\nA_1: send message\nA_2: receive/process message\nA_3: go offline\nA_4: come back online\nA_5: do nothing")
		if reached_consensus(nodes) {
			fmt.Println("reached consensus")
			break
		}
		node_num := -1
		print_and_scan(fmt.Sprintf("Node {N_0...N_%v}: N_", num_replicas-1), &node_num)
		if node_num < 0 || node_num >= num_replicas {
			fmt.Printf("Invalid node number, choose between 0 and %v\n", num_replicas - 1)
			break
		}
		for {
			exit := true
			action_num := -1
			print_and_scan("Action {A_0...A_5}: A_", &action_num)
			switch action_num {
				case 0:
					proposal_num := -1
					print_and_scan("proposal number: ", &proposal_num)
					nodes[node_num].ConstructPrepareMessage(proposal_num)
				case 1:
					if len(nodes[node_num].OutboundMessages) == 0 {
						fmt.Printf("Node N_%v does not have messages to send", node_num)
						break
					}
					m := nodes[node_num].OutboundMessages[0]
					receivers_string := ""
					print_and_scan("which nodes to send message to (comma seperated list of node numbers): ", &receivers_string)
					receivers_string_slice := strings.Split(receivers_string, ",")
					receivers := make([]int, len(receivers_string_slice))
					invalid_input := false
					for index, rs := range receivers_string_slice {
						r, err := strconv.Atoi(rs)
						if err != nil || r < 0 || r >= num_replicas {
							fmt.Printf("invalid node num %v, specify nodes between 0 and %v\n", rs, num_replicas - 1)
							invalid_input = true
							break
						}
						receivers[index] = r
					}
					if invalid_input {
						exit = false
						break
					}
					for _, r := range receivers {
						nodes[node_num].SendMessage(nodes[r], &m)
					}
					nodes[node_num].OutboundMessages = nodes[node_num].OutboundMessages[1:]
				case 2:
					m, err := nodes[node_num].ReceiveMessage()
					if err != nil {
						fmt.Printf("%v", err)
						break
					}
					if m == nil {
						fmt.Println("no messages to process")
						break
					}
					nodes[node_num].ProcessMessageAndConstructResponse(m)
				case 3: 
					nodes[node_num].Crash()
				case 4:
					nodes[node_num].Recover()
				case 5:
					break
				default:
					fmt.Println("invalid action number")
					exit = false
			}
			if exit {
				break
			}
		}

		/*
		fmt.Println("choose node: A, B, or C")
		var node_string string
		fmt.Scanln(&node_string)
		var n *node.Node 
		switch node_string {
			case "A":
				n = A
			case "B":
				n = B
			case "C":
				n = C
			default: 
				exit = true
		}
		if exit {
			break
		}
		fmt.Println("choose command:\n0: construct prepare message\n1: receive/process message\n2: send message\n3: clear outbound messages\n4: crash\n5: recover")
		var cmd int
		fmt.Scanln(&cmd)
		switch cmd {
			case 0:
				fmt.Println("specify proposal number")
				proposal_num := 0
				fmt.Scanln(&proposal_num)
				n.ConstructPrepareMessage(proposal_num)
			case 1:
				if len(n.MessageQueue) == 0 {
					fmt.Println("no messages to receive")
					break
				}
				m := n.MessageQueue[0]
				n.MessageQueue = n.MessageQueue[1:]
				n.ProcessMessageAndConstructResponse(&m)
			case 2:
				if len(n.OutboundMessages) == 0 {
					fmt.Println("no messages to send")
					break
				}
				m := n.OutboundMessages[0]
				fmt.Println("specify destination")
				var dest string
				fmt.Scanln(&dest)
				switch dest {
					case "A":
						n.SendMessage(A, &m)
					case "B":
						n.SendMessage(B, &m)
					case "C":
						n.SendMessage(C, &m)
					default:
						exit = true
				}
			case 3:
				n.OutboundMessages = n.OutboundMessages[1:]
			case 4:
				n.Crash()
			case 5:
				n.Recover()
			default:
				exit = true
		}
		if exit {
			break
		}
		*/
	}
}

func print_state(nodes []*node.Node) {
	for _, n := range nodes {
		n.Print()
	}
}

func reached_consensus(nodes []*node.Node) bool {
	counts := map[int]int{}
	max_count := 0
	for _, n := range nodes {
		if n.AcceptedProposal == nil {
			continue
		}
		if _, ok := counts[*n.AcceptedProposal]; !ok {
			counts[*n.AcceptedProposal] = 0
		}
		val := counts[*n.AcceptedProposal] + 1
		counts[*n.AcceptedProposal] = val 
		if val > max_count {
			max_count = val
		}
	}
	return max_count > len(nodes) / 2
}
