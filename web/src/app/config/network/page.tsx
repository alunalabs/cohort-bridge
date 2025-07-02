'use client';

import { Network } from 'lucide-react';
import FlexibleConfigBuilder from '@/components/FlexibleConfigBuilder';

export default function NetworkConfigPage() {
    const defaultValues = {
        database: {
            type: 'csv',
            filename: 'data/hospital_a_patients.csv',
            fields: ['name:FIRST', 'name:LAST', 'date:BIRTHDATE', 'gender:GENDER', 'zip:ZIP'],
            random_bits_percent: 0.02,
        },
        peer: {
            host: '10.0.1.100',
            port: 8082,
        },
        listen_port: 8081,

        security: {
            rate_limit_per_min: 5,
        },
        timeouts: {
            connection_timeout: '60s',
            read_timeout: '180s',
            write_timeout: '180s',
            idle_timeout: '900s',
            handshake_timeout: '90s',
        },
        matching: {
            bloom_size: 2048,
            bloom_hashes: 8,
            minhash_size: 256,
            qgram_length: 3,
            hamming_threshold: 200,
            jaccard_threshold: 0.75,
            qgram_threshold: 0.85,
            noise_level: 0.02,
        },
        logging: {
            level: 'info',
            file: 'logs/network_node.log',
            max_size: 200,
            max_backups: 5,
            max_age: 14,
            enable_audit: true,
            audit_file: 'logs/network_audit.log',
        },
    };

    return (
        <FlexibleConfigBuilder
            title="Network Configuration"
            description="Multi-party network setup with advanced matching algorithms and data normalization"
            defaultSections={['database', 'peer', 'security', 'timeouts', 'matching', 'logging']}
            defaultValues={defaultValues}
            icon={Network}
        />
    );
} 