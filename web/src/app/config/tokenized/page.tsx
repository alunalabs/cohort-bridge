'use client';

import { FileText } from 'lucide-react';
import FlexibleConfigBuilder from '@/components/FlexibleConfigBuilder';

export default function TokenizedConfigPage() {
    const defaultValues = {
        database: {
            type: 'csv',
            is_tokenized: true,
            tokenized_file: 'out/tokens_party_a.json',
            random_bits_percent: 0.0,
        },
        peer: {
            host: 'localhost',
            port: 8080,
        },
        listen_port: 8080,
    };

    return (
        <FlexibleConfigBuilder
            title="Tokenized Configuration"
            description="Work with pre-tokenized data files for enhanced privacy"
            defaultSections={['database', 'peer']}
            defaultValues={defaultValues}
            icon={FileText}
        />
    );
} 