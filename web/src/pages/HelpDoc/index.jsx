import React, { useMemo } from 'react';
import usageDocHtml from '../../assets/help/usage.html?raw';
import './HelpDoc.css';

const HelpDoc = () => {
  const html = useMemo(() => {
    return usageDocHtml
      .replaceAll('https://api.hanbbq.top', 'https://router.yeying.pub')
      .replace(
        /<p class="hero-subtitle"[^>]*>\s*API BaseURL（CF节点）：https:\/\/api\.aixhan\.com\s*<\/p>/g,
        '',
      )
      .replace(
        /<p class="hero-subtitle"[^>]*>\s*如果第一个节点无法使用就换另外一个CF节点\s*<\/p>/g,
        '',
      )
      .replace(
        /<pre><code class="language-[^"]*">([\s\S]*?)<\/code><\/pre>/g,
        (block, code) => {
          const normalizedCode = String(code || '');
          if (
            !/your-api-key-here|CCH_API_KEY|OPENAI_API_KEY|ANTHROPIC_AUTH_TOKEN|GOOGLE_API_KEY|REPLACE_WITH_RANDOM_TOKEN|\"token\":/i.test(
              normalizedCode,
            )
          ) {
            return block;
          }
          return `${block}<div class="router-help-doc-token-tip">需要令牌？<a href="/workspace/token">获取</a></div>`;
        },
      );
  }, []);

  return (
    <div className='dashboard-container'>
      <div
        className='router-help-doc-page'
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </div>
  );
};

export default HelpDoc;
