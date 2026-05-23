create index if not exists company_wiki_page_search_tsv_idx
  on company_wiki_page using gin (
    to_tsvector('english', coalesce(slug, '') || ' ' || coalesce(title, ''))
  );

create index if not exists company_wiki_revision_body_search_tsv_idx
  on company_wiki_revision using gin (
    to_tsvector('english', coalesce(body, ''))
  );
