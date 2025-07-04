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
        security: {
            rate_limit_per_min: 3,
        },
        matching: {
            hamming_threshold: 90,
            jaccard_threshold: 0.5,
        },
        timeouts: {
            connection_timeout: '30s',
            read_timeout: '60s',
            write_timeout: '60s',
            idle_timeout: '300s',
            handshake_timeout: '30s',
        },
        logging: {
            level: 'info',
            file: 'logs/cohort_tokenized.log',
            max_size: 100,
            max_backups: 3,
            max_age: 30,
            enable_audit: true,
            audit_file: 'logs/audit_tokenized.log',
        },
    };

    return (
        <FlexibleConfigBuilder
            title="Tokenized Configuration"
            description="Work with pre-tokenized data files for enhanced privacy"
            defaultSections={['database', 'peer', 'security', 'timeouts', 'logging']}
            defaultValues={defaultValues}
            icon={FileText}
        />
    );
} 