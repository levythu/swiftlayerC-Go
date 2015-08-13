// Split-tree is a data structure brainstormed by me (Although I think it has been invented by
// others and famous but I kept oblivious for a long time) to solve the labling problem that all
// the nodes on one segment tree should not change its lable during the process of adding nodes.
//
// The tree is like this:
//                             1000
//             0100                            1100
//     0010            0110            1010            1110
// 0001    0011    0101    0111    1001    1011    1101    1111        #Leaf
//   0       1       2       3       4       5       6       7         #Node
//
// layer(i)    =i&-i
// left(i)     =i-layer(i)>>1=i-(i&-i)>>1
// right(i)    =i+layer(i)>>1=i+(i&-i)>>1
// parent(i)   =(i^layer(i))|(layer(i)<<1)=(i^(i&-i))|((i&-i)<<1)
//
// By Levy

package splittree

// Leaf is the number in the tree, node is numbered from 0
func FromNodeToLeaf(n uint32) uint32 {
	return (n<<1)+1
}

func IsLeaf(i uint32) bool {
    return i&1==1
}

func FromLeaftoNode(i uint32) uint32 {
    return i>>1
}

func Parent(i uint32) uint32 {
    var layer=i&-i
    return (i^layer)|(layer<<1)
}

func Left(i uint32) uint32 {
    return i-((i&-i)>>1)
}

func Right(i uint32) uint32 {
    return i+((i&-i)>>1)
}

func GetRootLable(leafnum uint32) uint32 {
    leafnum=((leafnum-1)<<1)|1
    var res=leafnum
    for leafnum>0 {
        res=leafnum
        leafnum^=leafnum&-leafnum
    }
    return res
}
