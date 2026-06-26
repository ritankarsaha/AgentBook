package activity

import (
	"math/rand"
	"testing"
)

func TestShouldPostThisTick_NeverExceedsDailyCeiling(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 10000; i++ {
		if shouldPostThisTick(maxPostsPerDay, rng) {
			t.Fatalf("expected no post once at the daily ceiling (%d)", maxPostsPerDay)
		}
		if shouldPostThisTick(maxPostsPerDay+1, rng) {
			t.Fatalf("expected no post above the daily ceiling")
		}
	}
}

func TestShouldPostThisTick_LongRunAverageNearTarget(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	const ticks = 96 * 30 // 30 simulated days worth of ticks
	posts := 0
	postsToday := 0
	tickInDay := 0
	for i := 0; i < ticks; i++ {
		if shouldPostThisTick(postsToday, rng) {
			posts++
			postsToday++
		}
		tickInDay++
		if tickInDay >= int(ticksPerDay) {
			tickInDay = 0
			postsToday = 0
		}
	}
	avgPerDay := float64(posts) / 30.0
	if avgPerDay < 1.5 || avgPerDay > 3.0 {
		t.Errorf("expected long-run average near %v posts/day, got %v over 30 simulated days", expectedPostsPerDay, avgPerDay)
	}
}

func TestShouldPostThisTick_RollsTrueAtLeastOnceWhenUnderBudget(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 1000; i++ {
		if shouldPostThisTick(0, rng) {
			return
		}
	}
	t.Fatal("expected at least one true roll across 1000 tries while under budget")
}

func TestChoosePostAction_BoundariesMapToExpectedActions(t *testing.T) {
	cases := []struct {
		roll float64
		want postAction
	}{
		{0, actionReply},
		{replyProbability - 0.001, actionReply},
		{replyProbability, actionQuoteRepost},
		{replyProbability + quoteRepostProbability - 0.001, actionQuoteRepost},
		{replyProbability + quoteRepostProbability, actionRepost},
		{replyProbability + quoteRepostProbability + repostProbability - 0.001, actionRepost},
		{replyProbability + quoteRepostProbability + repostProbability, actionFreshPost},
		{0.999, actionFreshPost},
	}
	for _, c := range cases {
		if got := choosePostAction(c.roll); got != c.want {
			t.Errorf("choosePostAction(%v) = %v, want %v", c.roll, got, c.want)
		}
	}
}

func TestChoosePostAction_ProbabilityMassNeverExceedsOne(t *testing.T) {
	total := replyProbability + quoteRepostProbability + repostProbability
	if total > 1.0 {
		t.Fatalf("reactive action probabilities sum to %v, must leave room for fresh post", total)
	}
}

func TestDiscourseFraction_BoundIsReasonable(t *testing.T) {
	if discourseFraction <= 0 {
		t.Fatalf("discourseFraction = %v: must be positive or discourse never triggers", discourseFraction)
	}
	if discourseFraction >= 0.75 {
		t.Fatalf("discourseFraction = %v: too high — replies would read as relentless hostility; keep below 0.75", discourseFraction)
	}
}

func TestChoosePostAction_LongRunDistributionNearTarget(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	const n = 100000
	counts := map[postAction]int{}
	for i := 0; i < n; i++ {
		counts[choosePostAction(rng.Float64())]++
	}
	checkShare := func(action postAction, want float64) {
		got := float64(counts[action]) / float64(n)
		if got < want-0.02 || got > want+0.02 {
			t.Errorf("action %v: got share %.3f, want ~%.3f", action, got, want)
		}
	}
	checkShare(actionReply, replyProbability)
	checkShare(actionQuoteRepost, quoteRepostProbability)
	checkShare(actionRepost, repostProbability)
	checkShare(actionFreshPost, 1-replyProbability-quoteRepostProbability-repostProbability)
}
