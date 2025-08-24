package generator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/core-banking/perf-test/internal/config"
	"github.com/core-banking/perf-test/internal/executor"
	"github.com/core-banking/perf-test/internal/metrics"
)

type Generator struct {
	config         *config.Config
	scenario       *Scenario
	executor       *executor.Executor
	collector      *metrics.Collector
	accounts       []string
	stopChan       chan struct{}
	wg             sync.WaitGroup
	operationCount int64
	targetOps      int64
	stopOnce       sync.Once
}

func New(cfg *config.Config, scenario *Scenario, collector *metrics.Collector) *Generator {
	return &Generator{
		config:    cfg,
		scenario:  scenario,
		executor:  executor.New(cfg.APIURL),
		collector: collector,
		stopChan:  make(chan struct{}),
		targetOps: scenario.TargetOperations,
	}
}

func (g *Generator) Run(ctx context.Context) {
	log.Printf("Setting up %d accounts with initial balance %.2f", g.scenario.Accounts, g.scenario.InitialBalance)
	
	if err := g.setupAccounts(ctx); err != nil {
		log.Printf("Failed to setup accounts: %v", err)
		log.Printf("Continuing with existing accounts...")
		return
	}

	log.Printf("Starting load generation with %d workers", g.config.Workers)
	
	if g.config.RampUp > 0 {
		g.rampUp(ctx)
	} else {
		g.startWorkers(ctx, g.config.Workers)
	}

	<-ctx.Done()
	close(g.stopChan)
	g.wg.Wait()
}

func (g *Generator) setupAccounts(ctx context.Context) error {
	g.accounts = make([]string, 0, g.scenario.Accounts)
	
	setupStart := time.Now()
	var setupWg sync.WaitGroup
	accountChan := make(chan string, g.scenario.Accounts)
	errorChan := make(chan error, g.scenario.Accounts)
	
	concurrency := min(50, g.scenario.Accounts)
	semaphore := make(chan struct{}, concurrency)
	
	for i := 0; i < g.scenario.Accounts; i++ {
		setupWg.Add(1)
		go func(index int) {
			defer setupWg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			ownerName := fmt.Sprintf("perf-test-user-%d-%d", time.Now().Unix(), index)
			
			accountID, err := g.executor.CreateAccount(ctx, ownerName)
			if err != nil {
				errorChan <- fmt.Errorf("failed to create account %s: %w", ownerName, err)
				return
			}
			
			// Add initial balance if configured
			if g.scenario.InitialBalance > 0 {
				err = g.executor.Deposit(ctx, accountID, g.scenario.InitialBalance)
				if err != nil {
					errorChan <- fmt.Errorf("failed to deposit initial balance for account %s: %w", accountID, err)
					return
				}
			}
			
			accountChan <- accountID
		}(i)
	}
	
	go func() {
		setupWg.Wait()
		close(accountChan)
		close(errorChan)
	}()
	
	for accountID := range accountChan {
		g.accounts = append(g.accounts, accountID)
	}
	
	var errors []error
	for err := range errorChan {
		if err != nil {
			log.Printf("Account setup error: %v", err)
			errors = append(errors, err)
		}
	}
	
	if len(g.accounts) == 0 {
		return fmt.Errorf("failed to create any accounts, last errors: %v", errors)
	}
	
	if len(errors) > 0 {
		log.Printf("Account setup completed with %d errors, continuing with %d successful accounts", len(errors), len(g.accounts))
	}
	
	log.Printf("Created %d accounts in %.2fs", len(g.accounts), time.Since(setupStart).Seconds())
	return nil
}

func (g *Generator) rampUp(ctx context.Context) {
	rampUpSteps := min(10, g.config.Workers)
	if rampUpSteps == 0 {
		rampUpSteps = 1
	}
	
	stepDuration := g.config.RampUp / time.Duration(rampUpSteps)
	workersPerStep := max(1, g.config.Workers / rampUpSteps)
	
	workersStarted := 0
	for i := 1; i <= rampUpSteps; i++ {
		targetWorkers := min(workersPerStep * i, g.config.Workers)
		workersToStart := targetWorkers - workersStarted
		
		if workersToStart > 0 {
			log.Printf("Ramping up: %d/%d workers", targetWorkers, g.config.Workers)
			g.startWorkers(ctx, workersToStart)
			workersStarted = targetWorkers
		}
		
		if workersStarted >= g.config.Workers {
			break
		}
		
		select {
		case <-time.After(stepDuration):
		case <-ctx.Done():
			return
		}
	}
}

func (g *Generator) startWorkers(ctx context.Context, count int) {
	for i := 0; i < count; i++ {
		g.wg.Add(1)
		go g.worker(ctx, i)
	}
}

func (g *Generator) worker(ctx context.Context, id int) {
	defer g.wg.Done()
	
	for {
		// Check if we've reached the target operation count
		if atomic.LoadInt64(&g.operationCount) >= g.targetOps {
			return
		}
		
		select {
		case <-ctx.Done():
			return
		case <-g.stopChan:
			return
		default:
			operation := g.scenario.GenerateOperation(g.accounts)
			
			start := time.Now()
			err := g.executeOperation(ctx, operation)
			duration := time.Since(start)
			
			success := err == nil
			g.collector.RecordOperation(string(operation.Type), duration, success, err)
			
			// Increment global operation count and check if we've reached target
			newCount := atomic.AddInt64(&g.operationCount, 1)
			if newCount >= g.targetOps {
				log.Printf("Target operations reached: %d/%d - stopping worker", newCount, g.targetOps)
				g.stopOnce.Do(func() { 
					log.Printf("Closing stop channel - test should complete now")
					close(g.stopChan) 
				})
				return
			}
			
			if g.scenario.ThinkTime > 0 {
				time.Sleep(g.scenario.ThinkTime)
			}
		}
	}
}

func (g *Generator) executeOperation(ctx context.Context, op Operation) error {
	switch op.Type {
	case OpDeposit:
		return g.executor.Deposit(ctx, op.AccountID, op.Amount)
	case OpWithdraw:
		return g.executor.Withdraw(ctx, op.AccountID, op.Amount)
	case OpTransfer:
		return g.executor.Transfer(ctx, op.FromID, op.ToID, op.Amount)
	case OpBalance:
		_, err := g.executor.GetBalance(ctx, op.AccountID)
		return err
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}