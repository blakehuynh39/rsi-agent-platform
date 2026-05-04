package control

import (
	"context"
	"log"

	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
)

func RunCompanyWikiCompiler(ctx context.Context, cfg config.Config, repo any) error {
	result, err := companyknowledge.RunCompanyWikiCompiler(ctx, cfg, repo, nil)
	if err != nil {
		return err
	}
	log.Printf("company-wiki-compiler ok=%t compiler_run_id=%s claimed=%d published_pages=%d failed_items=%v",
		result.OK, result.CompilerRunID, result.Claimed, result.PublishedPages, result.FailedItems)
	return nil
}
