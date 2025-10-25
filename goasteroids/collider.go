package goasteroids

import "github.com/solarlune/resolv"

func (g *GameScene) checkCollision(obj, against *resolv.Circle) bool {
	// Test if obj intersects with anything on touching cell
	if against == nil {
		return obj.IntersectionTest(resolv.IntersectionTestSettings{
			TestAgainst: obj.SelectTouchingCells(1).FilterShapes(),
			OnIntersect: func(set resolv.IntersectionSet) bool {
				return true
			},
		})
	}

	// Test if obj intersects with specific object  (the against)
	return obj.IntersectionTest(resolv.IntersectionTestSettings{
		TestAgainst: against.SelectTouchingCells(1).FilterShapes(),
		OnIntersect: func(set resolv.IntersectionSet) bool {
			return true
		},
	})
}
