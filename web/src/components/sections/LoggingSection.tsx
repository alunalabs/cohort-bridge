'use client';

import { useFormContext } from 'react-hook-form';
import { FileText } from 'lucide-react';

export default function LoggingSection() {
    const { register } = useFormContext();

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
                        <FileText className="h-4 w-4 text-purple-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Logging Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure detailed logging for monitoring, debugging, and security auditing purposes.
                </p>
            </div>

            <div className="space-y-6">
                {/* Log Level */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Log Level
                    </label>
                    <select
                        {...register('logging.level')}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    >
                        <option value="debug">Debug</option>
                        <option value="info">Info</option>
                        <option value="warn">Warning</option>
                        <option value="error">Error</option>
                    </select>
                    <p className="mt-1 text-xs text-slate-500">
                        Minimum log level to capture
                    </p>
                </div>

                {/* Log File */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Log File Path
                    </label>
                    <input
                        type="text"
                        {...register('logging.log_file')}
                        placeholder="logs/cohort_bridge.log"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Path where log files will be written. Directory must exist and be writable
                    </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    {/* Max File Size */}
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Max File Size (MB)
                        </label>
                        <input
                            type="number"
                            {...register('logging.max_file_size', { valueAsNumber: true })}
                            placeholder="100"
                            min="1"
                            max="1000"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Maximum size before log rotation. Prevents disk space exhaustion
                        </p>
                    </div>

                    {/* Max Backup Files */}
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Max Backup Files
                        </label>
                        <input
                            type="number"
                            {...register('logging.max_backup_files', { valueAsNumber: true })}
                            placeholder="5"
                            min="0"
                            max="50"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Number of rotated log files to keep. 0 disables rotation
                        </p>
                    </div>

                    {/* Max Age */}
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Max Age (days)
                        </label>
                        <input
                            type="number"
                            {...register('logging.max_age_days', { valueAsNumber: true })}
                            placeholder="30"
                            min="1"
                            max="365"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Delete log files older than this many days
                        </p>
                    </div>
                </div>

                {/* Compression */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Compression Format
                    </label>
                    <input
                        type="text"
                        {...register('logging.compression')}
                        placeholder="gzip"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Compression for rotated logs: gzip, zip, or empty for none
                    </p>
                </div>

                {/* Enable Syslog */}
                <div>
                    <label className="flex items-center space-x-3">
                        <input
                            type="checkbox"
                            {...register('logging.enable_syslog')}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm font-medium text-slate-700">Enable Syslog Output</span>
                    </label>
                    <p className="mt-1 text-xs text-slate-500 ml-6">
                        Send logs to system syslog (disabled for Windows compatibility)
                    </p>
                </div>

                {/* Enable Audit */}
                <div>
                    <label className="flex items-center space-x-3">
                        <input
                            type="checkbox"
                            {...register('logging.enable_audit')}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm font-medium text-slate-700">Enable Security Audit Logging</span>
                    </label>
                    <p className="mt-1 text-xs text-slate-500 ml-6">
                        Enable detailed security event logging
                    </p>
                </div>

                {/* Audit Log File */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Audit Log File
                    </label>
                    <input
                        type="text"
                        {...register('logging.audit_log_file')}
                        placeholder="logs/audit.log"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Separate file for security and compliance audit trails
                    </p>
                </div>
            </div>
        </div>
    );
} 