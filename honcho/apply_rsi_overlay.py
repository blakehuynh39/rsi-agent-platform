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
    replace_once(
        clients_path,
        "def convert_tools_for_provider(\n"
        "    tools: list[dict[str, Any]],\n"
        "    provider: SupportedProviders,\n"
        ") -> list[dict[str, Any]]:\n"
        '    """\n'
        "    Convert tool definitions to provider-specific format.\n"
        "\n"
        "    Args:\n"
        "        tools: List of tool definitions in Anthropic format (with input_schema)\n"
        "        provider: The target provider to convert tools for\n"
        "\n"
        "    Returns:\n"
        "        List of tool definitions in the provider's native format\n"
        '    """\n'
        '    if provider == "anthropic":\n'
        "        # Anthropic format: input_schema\n"
        "        return tools\n"
        '    elif provider in ("openai", "custom", "vllm"):\n'
        "        # OpenAI format: parameters instead of input_schema\n"
        "        # custom and vllm use AsyncOpenAI client so need OpenAI format\n"
        "        return [\n"
        "            {\n"
        '                "type": "function",\n'
        '                "function": {\n'
        '                    "name": tool["name"],\n'
        '                    "description": tool["description"],\n'
        '                    "parameters": tool["input_schema"],\n'
        "                },\n"
        "            }\n"
        "            for tool in tools\n"
        "        ]\n"
        '    elif provider == "google":\n'
        "        # Google format: function_declarations wrapped in a tool object\n"
        "        return [\n"
        "            {\n"
        '                "function_declarations": [\n'
        "                    {\n"
        '                        "name": tool["name"],\n'
        '                        "description": tool["description"],\n'
        '                        "parameters": tool["input_schema"],\n'
        "                    }\n"
        "                    for tool in tools\n"
        "                ]\n"
        "            }\n"
        "        ]\n"
        "    else:\n"
        "        # For unsupported providers, return as-is (will likely error if tools are used)\n"
        "        logger.warning(\n"
        '            f"Tool calling not implemented for provider {provider}, returning tools as-is"\n'
        "        )\n"
        "        return tools\n",
        "def convert_tools_for_provider(\n"
        "    tools: list[dict[str, Any]],\n"
        "    provider: SupportedProviders,\n"
        ") -> list[dict[str, Any]]:\n"
        '    """\n'
        "    Convert tool definitions to provider-specific format.\n"
        "\n"
        "    Args:\n"
        "        tools: List of tool definitions in Anthropic format (with input_schema)\n"
        "        provider: The target provider to convert tools for\n"
        "\n"
        "    Returns:\n"
        "        List of tool definitions in the provider's native format\n"
        '    """\n'
        '    if provider == "anthropic":\n'
        "        # Anthropic format: input_schema\n"
        "        return tools\n"
        '    elif provider in ("openai", "custom", "vllm"):\n'
        "        # OpenAI format: parameters instead of input_schema\n"
        "        # custom and vllm use AsyncOpenAI client so need OpenAI format\n"
        "        return [\n"
        "            {\n"
        '                "type": "function",\n'
        '                "function": {\n'
        '                    "name": tool["name"],\n'
        '                    "description": tool["description"],\n'
        '                    "parameters": tool["input_schema"],\n'
        "                },\n"
        "            }\n"
        "            for tool in tools\n"
        "        ]\n"
        '    elif provider == "google":\n'
        "        # Google format: function_declarations wrapped in a tool object\n"
        "        return [\n"
        "            {\n"
        '                "function_declarations": [\n'
        "                    {\n"
        '                        "name": tool["name"],\n'
        '                        "description": tool["description"],\n'
        '                        "parameters": tool["input_schema"],\n'
        "                    }\n"
        "                    for tool in tools\n"
        "                ]\n"
        "            }\n"
        "        ]\n"
        "    else:\n"
        "        # For unsupported providers, return as-is (will likely error if tools are used)\n"
        "        logger.warning(\n"
        '            f"Tool calling not implemented for provider {provider}, returning tools as-is"\n'
        "        )\n"
        "        return tools\n"
        "\n"
        "\n"
        "def _should_use_openai_responses_api(\n"
        "    model: str,\n"
        "    *,\n"
        "    tools: list[dict[str, Any]] | None,\n"
        "    reasoning_effort: ReasoningEffortType,\n"
        "    verbosity: VerbosityType,\n"
        "    response_model: type[BaseModel] | None,\n"
        "    json_mode: bool,\n"
        "    stream: bool,\n"
        ") -> bool:\n"
        "    if stream:\n"
        "        return False\n"
        '    if "gpt-5" not in model:\n'
        "        return False\n"
        "    return bool(tools or reasoning_effort or verbosity or response_model or json_mode)\n"
        "\n"
        "\n"
        "def _openai_message_content_to_string(content: Any) -> str:\n"
        "    if isinstance(content, str):\n"
        "        return content\n"
        "    if isinstance(content, list):\n"
        "        parts: list[str] = []\n"
        "        for item in content:\n"
        "            if isinstance(item, dict):\n"
        '                text = item.get("text")\n'
        "                if isinstance(text, str) and text:\n"
        "                    parts.append(text)\n"
        "        if parts:\n"
        '            return "\\n".join(parts)\n'
        "        return json.dumps(content)\n"
        "    if content is None:\n"
        '        return ""\n'
        "    return str(content)\n"
        "\n"
        "\n"
        "def _convert_openai_messages_to_responses_input(\n"
        "    messages: list[dict[str, Any]],\n"
        ") -> list[dict[str, Any]]:\n"
        "    input_items: list[dict[str, Any]] = []\n"
        "    for msg in messages:\n"
        '        role = str(msg.get("role", "") or "").strip()\n'
        '        if role == "tool":\n'
        "            input_items.append(\n"
        "                {\n"
        '                    "type": "function_call_output",\n'
        '                    "call_id": str(msg.get("tool_call_id", "") or "").strip(),\n'
        '                    "output": _openai_message_content_to_string(msg.get("content")),\n'
        '                    "id": str(msg.get("tool_call_id", "") or "").strip(),\n'
        "                }\n"
        "            )\n"
        "            continue\n"
        "\n"
        '        content = _openai_message_content_to_string(msg.get("content"))\n'
        '        if role in {"system", "developer", "user", "assistant"} and content:\n'
        '            input_items.append({"role": role, "content": content})\n'
        "\n"
        '        tool_calls = msg.get("tool_calls")\n'
        '        if role == "assistant" and isinstance(tool_calls, list):\n'
        "            for tool_call in tool_calls:\n"
        "                if not isinstance(tool_call, dict):\n"
        "                    continue\n"
        '                function = tool_call.get("function", {})\n'
        "                if not isinstance(function, dict):\n"
        "                    continue\n"
        '                call_id = str(tool_call.get("id", "") or "").strip()\n'
        '                name = str(function.get("name", "") or "").strip()\n'
        '                arguments = function.get("arguments")\n'
        "                if not name or arguments is None:\n"
        "                    continue\n"
        "                input_items.append(\n"
        "                    {\n"
        '                        "type": "function_call",\n'
        '                        "call_id": call_id,\n'
        '                        "id": call_id,\n'
        '                        "name": name,\n'
        '                        "arguments": str(arguments),\n'
        "                    }\n"
        "                )\n"
        "    return input_items\n"
        "\n"
        "\n"
        "def _convert_openai_tools_to_responses_tools(\n"
        "    tools: list[dict[str, Any]],\n"
        ") -> list[dict[str, Any]]:\n"
        "    converted: list[dict[str, Any]] = []\n"
        "    for tool in tools:\n"
        "        if not isinstance(tool, dict):\n"
        "            continue\n"
        '        if tool.get("type") != "function":\n'
        "            continue\n"
        '        function = tool.get("function", {})\n'
        "        if not isinstance(function, dict):\n"
        "            continue\n"
        '        name = str(function.get("name", "") or "").strip()\n'
        "        if not name:\n"
        "            continue\n"
        "        converted.append(\n"
        "            {\n"
        '                "type": "function",\n'
        '                "name": name,\n'
        '                "description": function.get("description"),\n'
        '                "parameters": function.get("parameters") or {},\n'
        "            }\n"
        "        )\n"
        "    return converted\n"
        "\n"
        "\n"
        "def _convert_openai_tool_choice_to_responses(\n"
        "    tool_choice: str | dict[str, Any] | None,\n"
        ") -> str | dict[str, Any] | None:\n"
        "    if tool_choice is None:\n"
        "        return None\n"
        "    if isinstance(tool_choice, str):\n"
        "        return tool_choice\n"
        '    if "name" in tool_choice:\n'
        '        return {"type": "function", "name": tool_choice["name"]}\n'
        '    function = tool_choice.get("function")\n'
        '    if isinstance(function, dict) and "name" in function:\n'
        '        return {"type": "function", "name": function["name"]}\n'
        "    return tool_choice\n"
        "\n"
        "\n"
        "def _extract_openai_responses_tool_calls(response: Any) -> list[dict[str, Any]]:\n"
        "    tool_calls: list[dict[str, Any]] = []\n"
        '    for item in getattr(response, "output", []) or []:\n'
        '        if getattr(item, "type", None) != "function_call":\n'
        "            continue\n"
        '        arguments = getattr(item, "arguments", "") or ""\n'
        "        try:\n"
        "            parsed_arguments = json.loads(arguments) if arguments else {}\n"
        "        except json.JSONDecodeError:\n"
        "            parsed_arguments = {}\n"
        "        tool_calls.append(\n"
        "            {\n"
        '                "id": getattr(item, "call_id", "") or getattr(item, "id", ""),\n'
        '                "name": getattr(item, "name", ""),\n'
        '                "input": parsed_arguments,\n'
        "            }\n"
        "        )\n"
        "    return tool_calls\n"
        "\n"
        "\n"
        "def _extract_openai_responses_reasoning_content(response: Any) -> str | None:\n"
        "    parts: list[str] = []\n"
        '    for item in getattr(response, "output", []) or []:\n'
        '        if getattr(item, "type", None) != "reasoning":\n'
        "            continue\n"
        '        summaries = getattr(item, "summary", None) or []\n'
        "        for summary in summaries:\n"
        '            text = getattr(summary, "text", None)\n'
        "            if text:\n"
        "                parts.append(text)\n"
        '    return "\\n".join(parts) if parts else None\n'
        "\n"
        "\n"
        "def _responses_text_format(\n"
        "    response_model: type[BaseModel] | None,\n"
        "    json_mode: bool,\n"
        ") -> dict[str, Any] | None:\n"
        "    if response_model is not None:\n"
        "        return {\n"
        '            "format": {\n'
        '                "type": "json_schema",\n'
        '                "name": response_model.__name__,\n'
        '                "schema": response_model.model_json_schema(),\n'
        '                "strict": True,\n'
        "            }\n"
        "        }\n"
        "    if json_mode:\n"
        '        return {"format": {"type": "json_object"}}\n'
        "    return None\n"
        "\n"
        "\n"
        "async def _call_openai_responses_api(\n"
        "    client: AsyncOpenAI,\n"
        "    *,\n"
        "    model: str,\n"
        "    messages: list[dict[str, Any]],\n"
        "    max_tokens: int,\n"
        "    response_model: type[BaseModel] | None,\n"
        "    json_mode: bool,\n"
        "    temperature: float | None,\n"
        "    stop_seqs: list[str] | None,\n"
        "    reasoning_effort: ReasoningEffortType,\n"
        "    verbosity: VerbosityType,\n"
        "    tools: list[dict[str, Any]] | None,\n"
        "    tool_choice: str | dict[str, Any] | None,\n"
        ') -> "HonchoLLMCallResponse[Any]":\n'
        "    if stop_seqs:\n"
        '        logger.warning("OpenAI Responses API path ignores stop sequences")\n'
        "    input_items = _convert_openai_messages_to_responses_input(messages)\n"
        "    response_text = _responses_text_format(response_model, json_mode)\n"
        "    if response_text is not None and verbosity:\n"
        '        response_text["verbosity"] = verbosity\n'
        "    elif verbosity:\n"
        '        response_text = {"verbosity": verbosity}\n'
        "\n"
        "    create_params: dict[str, Any] = {\n"
        '        "model": model,\n'
        '        "input": input_items,\n'
        '        "max_output_tokens": max_tokens,\n'
        "    }\n"
        '    if temperature is not None and "gpt-5" not in model:\n'
        '        create_params["temperature"] = temperature\n'
        "    if reasoning_effort:\n"
        '        create_params["reasoning"] = {"effort": reasoning_effort}\n'
        "    if response_text is not None:\n"
        '        create_params["text"] = response_text\n'
        "    if tools:\n"
        '        create_params["tools"] = _convert_openai_tools_to_responses_tools(tools)\n'
        "        converted_tool_choice = _convert_openai_tool_choice_to_responses(tool_choice)\n"
        "        if converted_tool_choice is not None:\n"
        '            create_params["tool_choice"] = converted_tool_choice\n'
        "\n"
        "    response = await client.responses.create(**create_params)\n"
        '    usage = getattr(response, "usage", None)\n'
        "    tool_calls = _extract_openai_responses_tool_calls(response)\n"
        '    content_text = getattr(response, "output_text", "") or ""\n'
        "\n"
        "    parsed_content: Any = content_text\n"
        "    if response_model is not None:\n"
        "        parsed_json = json.loads(content_text)\n"
        "        parsed_content = response_model.model_validate(parsed_json)\n"
        "\n"
        "    cache_creation, cache_read = extract_openai_cache_tokens(usage)\n"
        '    finish_reason = getattr(response, "status", None)\n'
        "    return HonchoLLMCallResponse(\n"
        "        content=parsed_content,\n"
        '        input_tokens=getattr(usage, "input_tokens", 0) or 0,\n'
        '        output_tokens=getattr(usage, "output_tokens", 0) or 0,\n'
        "        cache_creation_input_tokens=cache_creation,\n"
        "        cache_read_input_tokens=cache_read,\n"
        "        finish_reasons=[finish_reason] if finish_reason else [],\n"
        "        tool_calls_made=tool_calls,\n"
        "        thinking_content=_extract_openai_responses_reasoning_content(response),\n"
        "    )\n",
    )
    replace_once(
        clients_path,
        "    # OpenAI native: usage.prompt_tokens_details.cached_tokens\n"
        "    if hasattr(usage, \"prompt_tokens_details\") and usage.prompt_tokens_details:\n"
        "        details = usage.prompt_tokens_details\n"
        "        if hasattr(details, \"cached_tokens\") and details.cached_tokens:\n"
        "            cache_read = details.cached_tokens\n"
        "\n"
        "    # OpenRouter style: usage.cache_read_input_tokens or usage.cached_tokens\n",
        "    # OpenAI native: usage.prompt_tokens_details.cached_tokens\n"
        "    if hasattr(usage, \"prompt_tokens_details\") and usage.prompt_tokens_details:\n"
        "        details = usage.prompt_tokens_details\n"
        "        if hasattr(details, \"cached_tokens\") and details.cached_tokens:\n"
        "            cache_read = details.cached_tokens\n"
        "\n"
        "    # OpenAI Responses API: usage.input_tokens_details.cached_tokens\n"
        "    if cache_read == 0 and hasattr(usage, \"input_tokens_details\") and usage.input_tokens_details:\n"
        "        details = usage.input_tokens_details\n"
        "        if hasattr(details, \"cached_tokens\") and details.cached_tokens:\n"
        "            cache_read = details.cached_tokens\n"
        "\n"
        "    # OpenRouter style: usage.cache_read_input_tokens or usage.cached_tokens\n",
    )
    replace_once(
        clients_path,
        "            openai_params: dict[str, Any] = {\n"
        '                "model": params["model"],\n'
        '                "messages": processed_messages,\n'
        "            }\n",
        "            if _should_use_openai_responses_api(\n"
        "                model,\n"
        "                tools=tools,\n"
        "                reasoning_effort=reasoning_effort,\n"
        "                verbosity=verbosity,\n"
        "                response_model=response_model,\n"
        "                json_mode=json_mode,\n"
        "                stream=stream,\n"
        "            ):\n"
        "                return await _call_openai_responses_api(\n"
        "                    client,\n"
        '                    model=params["model"],\n'
        "                    messages=processed_messages,\n"
        '                    max_tokens=params["max_tokens"],\n'
        "                    response_model=response_model,\n"
        "                    json_mode=json_mode,\n"
        "                    temperature=temperature,\n"
        "                    stop_seqs=stop_seqs,\n"
        "                    reasoning_effort=reasoning_effort,\n"
        "                    verbosity=verbosity,\n"
        "                    tools=tools,\n"
        "                    tool_choice=tool_choice,\n"
        "                )\n"
        "\n"
        "            openai_params: dict[str, Any] = {\n"
        '                "model": params["model"],\n'
        '                "messages": processed_messages,\n'
        "            }\n",
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
