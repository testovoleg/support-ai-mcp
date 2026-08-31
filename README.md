# support-ai-mcp

MCP-сервер для поиска подсказок, статей базы знаний и похожих заявок через API 5Systems Support AI.

## Настройка Claude Desktop

1. Склонируйте репозиторий к себе на компьютер и запомните путь до файла `support-ai-mcp.exe` — он уже собран, дополнительно ничего компилировать не нужно.

2. Откройте файл конфигурации Claude Desktop:
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`

3. Добавьте сервер в секцию `mcpServers`, указав полный путь до `support-ai-mcp.exe` из склонированного репозитория:

   ```json
   {
     "mcpServers": {
       "support-ai-mcp": {
         "command": "C:\\path\\to\\support-ai-mcp\\support-ai-mcp.exe"
       }
     }
   }
   ```

4. Перезапустите Claude Desktop. После перезапуска в списке инструментов должен появиться `support-ai-mcp` с инструментами `search_tips`, `search_articles`.
