#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path


def replace_once(path: Path, old: str, new: str) -> None:
    text = path.read_text(encoding="utf-8")
    if old not in text:
        raise SystemExit(f"expected snippet not found in {path}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")


def replace_all(path: Path, old: str, new: str) -> None:
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count == 0:
        raise SystemExit(f"expected snippet not found in {path}")
    path.write_text(text.replace(old, new), encoding="utf-8")


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit("usage: apply_rsi_overlay.py <honcho-root>")
    root = Path(sys.argv[1]).resolve()
    if not (root / "src").exists():
        raise SystemExit(f"{root} does not look like a Honcho checkout")

    config_path = root / "src/config.py"
    replace_once(
        config_path,
        '    PROVIDER: SupportedProviders = "google"\n'
        '    MODEL: str = "gemini-2.5-flash-lite"\n'
        "    TEMPERATURE: float | None = None\n",
        '    PROVIDER: SupportedProviders = "openai"\n'
        '    MODEL: str = "gpt-5.4"\n'
        '    REASONING_EFFORT: Literal["minimal", "low", "medium", "high", "xhigh"] | None = "xhigh"\n'
        "    TEMPERATURE: float | None = None\n",
    )
    replace_once(
        config_path,
        '    PROVIDER: Annotated[SupportedProviders, Field(validation_alias="provider")]\n'
        '    MODEL: Annotated[str, Field(validation_alias="model")]\n',
        '    PROVIDER: Annotated[SupportedProviders, Field(validation_alias="provider")]\n'
        '    MODEL: Annotated[str, Field(validation_alias="model")]\n'
        '    REASONING_EFFORT: Annotated[\n'
        '        Literal["minimal", "low", "medium", "high", "xhigh"] | None,\n'
        '        Field(default="xhigh", validation_alias="reasoning_effort"),\n'
        '    ] = "xhigh"\n',
    )
    replace_once(
        config_path,
        '            "minimal": DialecticLevelSettings(\n'
        '                PROVIDER="google",\n'
        '                MODEL="gemini-2.5-flash-lite",\n',
        '            "minimal": DialecticLevelSettings(\n'
        '                PROVIDER="openai",\n'
        '                MODEL="gpt-5.4",\n'
        '                REASONING_EFFORT="xhigh",\n',
    )
    replace_once(
        config_path,
        '            "low": DialecticLevelSettings(\n'
        '                PROVIDER="google",\n'
        '                MODEL="gemini-2.5-flash-lite",\n',
        '            "low": DialecticLevelSettings(\n'
        '                PROVIDER="openai",\n'
        '                MODEL="gpt-5.4",\n'
        '                REASONING_EFFORT="xhigh",\n',
    )
    replace_once(
        config_path,
        '            "medium": DialecticLevelSettings(\n'
        '                PROVIDER="anthropic",\n'
        '                MODEL="claude-haiku-4-5",\n',
        '            "medium": DialecticLevelSettings(\n'
        '                PROVIDER="openai",\n'
        '                MODEL="gpt-5.4",\n'
        '                REASONING_EFFORT="xhigh",\n',
    )
    replace_once(
        config_path,
        '            "high": DialecticLevelSettings(\n'
        '                PROVIDER="anthropic",\n'
        '                MODEL="claude-haiku-4-5",\n',
        '            "high": DialecticLevelSettings(\n'
        '                PROVIDER="openai",\n'
        '                MODEL="gpt-5.4",\n'
        '                REASONING_EFFORT="xhigh",\n',
    )
    replace_once(
        config_path,
        '            "max": DialecticLevelSettings(\n'
        '                PROVIDER="anthropic",\n'
        '                MODEL="claude-haiku-4-5",\n',
        '            "max": DialecticLevelSettings(\n'
        '                PROVIDER="openai",\n'
        '                MODEL="gpt-5.4",\n'
        '                REASONING_EFFORT="xhigh",\n',
    )
    replace_once(
        config_path,
        '    PROVIDER: SupportedProviders = "google"\n'
        '    MODEL: str = "gemini-2.5-flash"\n'
        "    MAX_TOKENS_SHORT: Annotated[int, Field(default=1000, gt=0, le=10_000)] = 1000\n",
        '    PROVIDER: SupportedProviders = "openai"\n'
        '    MODEL: str = "gpt-5.4"\n'
        '    REASONING_EFFORT: Literal["minimal", "low", "medium", "high", "xhigh"] | None = "xhigh"\n'
        "    MAX_TOKENS_SHORT: Annotated[int, Field(default=1000, gt=0, le=10_000)] = 1000\n",
    )

    clients_path = root / "src/utils/clients.py"
    replace_all(
        clients_path,
        'Literal["low", "medium", "high", "minimal"]',
        'Literal["low", "medium", "high", "minimal", "xhigh"]',
    )

    deriver_path = root / "src/deriver/deriver.py"
    replace_once(
        deriver_path,
        '        reasoning_effort="minimal",\n',
        "        reasoning_effort=settings.DERIVER.REASONING_EFFORT,\n",
    )

    summarizer_path = root / "src/utils/summarizer.py"
    replace_all(
        summarizer_path,
        "        max_tokens=settings.SUMMARY.MAX_TOKENS_SHORT,\n    )\n",
        "        max_tokens=settings.SUMMARY.MAX_TOKENS_SHORT,\n        reasoning_effort=settings.SUMMARY.REASONING_EFFORT,\n    )\n",
    )
    replace_all(
        summarizer_path,
        "        max_tokens=settings.SUMMARY.MAX_TOKENS_LONG,\n    )\n",
        "        max_tokens=settings.SUMMARY.MAX_TOKENS_LONG,\n        reasoning_effort=settings.SUMMARY.REASONING_EFFORT,\n    )\n",
    )

    dialectic_path = root / "src/dialectic/core.py"
    replace_all(
        dialectic_path,
        "            max_input_tokens=settings.DIALECTIC.MAX_INPUT_TOKENS,\n            trace_name=\"dialectic_chat\",\n",
        "            max_input_tokens=settings.DIALECTIC.MAX_INPUT_TOKENS,\n            reasoning_effort=level_settings.REASONING_EFFORT,\n            trace_name=\"dialectic_chat\",\n",
    )

    main_path = root / "src/main.py"
    replace_once(
        main_path,
        '@app.get("/health")\n'
        "async def health_check():\n"
        '    """Health check endpoint for monitoring and container orchestration."""\n'
        '    return {"status": "ok"}\n',
        'def honcho_runtime_payload() -> dict[str, object]:\n'
        "    return {\n"
        '        "namespace": settings.NAMESPACE,\n'
        '        "db_schema": settings.DB.SCHEMA,\n'
        '        "cache_enabled": settings.CACHE.ENABLED,\n'
        '        "cache_url_configured": bool(settings.CACHE.URL),\n'
        '        "deriver": {\n'
        '            "provider": settings.DERIVER.PROVIDER,\n'
        '            "model": settings.DERIVER.MODEL,\n'
        '            "reasoning_effort": settings.DERIVER.REASONING_EFFORT,\n'
        '        },\n'
        '        "summary": {\n'
        '            "provider": settings.SUMMARY.PROVIDER,\n'
        '            "model": settings.SUMMARY.MODEL,\n'
        '            "reasoning_effort": settings.SUMMARY.REASONING_EFFORT,\n'
        '        },\n'
        '        "dialectic_levels": {\n'
        '            level: {\n'
        '                "provider": level_settings.PROVIDER,\n'
        '                "model": level_settings.MODEL,\n'
        '                "reasoning_effort": level_settings.REASONING_EFFORT,\n'
        '                "thinking_budget_tokens": level_settings.THINKING_BUDGET_TOKENS,\n'
        '            }\n'
        '            for level, level_settings in settings.DIALECTIC.LEVELS.items()\n'
        '        },\n'
        "    }\n\n"
        '@app.get("/health")\n'
        "async def health_check():\n"
        '    """Health check endpoint for monitoring and container orchestration."""\n'
        '    return {"status": "ok", "runtime": honcho_runtime_payload()}\n\n'
        '@app.get("/runtimez")\n'
        "async def runtime_status():\n"
        '    payload = honcho_runtime_payload()\n'
        '    payload["status"] = "ok"\n'
        "    return payload\n",
    )


if __name__ == "__main__":
    main()
