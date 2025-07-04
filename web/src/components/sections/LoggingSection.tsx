'use client';

import { useFormContext } from 'react-hook-form';
import { FileText, Shield, Monitor } from 'lucide-react';

interface LoggingSectionProps {
    missingFields?: string[];
}

export default function LoggingSection({ missingFields = [] }: LoggingSectionProps) {
    const { register, setValue, watch } = useFormContext();

    const logLevel = watch('logging.level');
    const enableAudit = watch('logging.enable_audit');

    const getInputClass = (fieldName: string) => {
        const baseClass = "w-full px-3 py-2 border rounded-lg focus:ring-2 text-slate-900 placeholder-slate-400 bg-white transition-colors";
        const isMissing = missingFields.includes(fieldName);

        if (isMissing) {
            return `${baseClass} border-red-300 bg-red-50 focus:ring-red-500 focus:border-red-500`;
        }

        return `${baseClass} border-slate-300 focus:ring-blue-500 focus:border-blue-500`;
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
                        <FileText className="h-4 w-4 text-indigo-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Logging Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure logging levels, file rotation, and audit trails for monitoring and compliance.
                </p>
            </div>

            <div className="space-y-6">
                {/* Log Level */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-3">
                        Log Level
                    </label>
                    <div className="grid grid-cols-2 gap-3">
                        {['info', 'debug', 'warn', 'error'].map((level) => (
                            <button
                                key={level}
                                type="button"
                                onClick={() => setValue('logging.level', level)}
                                className={`p-3 rounded-lg border-2 transition-all duration-200 cursor-pointer ${logLevel === level
                                    ? 'border-blue-500 bg-blue-50 text-blue-700'
                                    : 'border-slate-300 hover:border-slate-400 text-slate-700'
                                    }`}
                            >
                                <div className="text-center">
                                    <div className="text-sm font-semibold capitalize">{level}</div>
                                    <div className="text-xs text-slate-500 mt-1">
                                        {level === 'info' && 'General information'}
                                        {level === 'debug' && 'Detailed debugging'}
                                        {level === 'warn' && 'Warning messages'}
                                        {level === 'error' && 'Error messages only'}
                                    </div>
                                </div>
                            </button>
                        ))}
                    </div>
                    <p className="mt-2 text-xs text-slate-600">
                        <strong>Info:</strong> Standard operations, <strong>Debug:</strong> Detailed tracing,
                        <strong>Warn:</strong> Potential issues, <strong>Error:</strong> Critical problems only
                    </p>
                </div>

                {/* Log File Configuration */}
                <div className="border border-slate-200 rounded-lg p-4 bg-slate-50">
                    <div className="flex items-center space-x-2 mb-4">
                        <FileText className="h-4 w-4 text-slate-600" />
                        <h4 className="text-sm font-medium text-slate-700">File Logging</h4>
                    </div>

                    <div className="space-y-4">
                        {/* Log File Path */}
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Log File Path
                            </label>
                            <input
                                type="text"
                                {...register('logging.file_path')}
                                placeholder="logs/cohort-bridge.log"
                                className={getInputClass('logging.file_path')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Path where log files will be written. Directory will be created if it doesn't exist.
                            </p>
                        </div>

                        {/* File Rotation Settings */}
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Max File Size (MB)
                                </label>
                                <input
                                    type="number"
                                    {...register('logging.max_file_size_mb', { valueAsNumber: true })}
                                    placeholder="100"
                                    min="1"
                                    max="1000"
                                    className={getInputClass('logging.max_file_size_mb')}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Max Backup Files
                                </label>
                                <input
                                    type="number"
                                    {...register('logging.max_backup_files', { valueAsNumber: true })}
                                    placeholder="3"
                                    min="1"
                                    max="50"
                                    className={getInputClass('logging.max_backup_files')}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Max Age (days)
                                </label>
                                <input
                                    type="number"
                                    {...register('logging.max_age', { valueAsNumber: true })}
                                    placeholder="30"
                                    min="1"
                                    max="365"
                                    className={getInputClass('logging.max_age')}
                                />
                            </div>
                        </div>
                    </div>
                </div>

                {/* Output Options */}
                <div className="space-y-4">
                    <h4 className="text-sm font-medium text-slate-700">Output Options</h4>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {/* Console Logging */}
                        <div className="border border-slate-200 rounded-lg p-3">
                            <label className="flex items-center space-x-3">
                                <input
                                    type="checkbox"
                                    {...register('logging.console_enabled')}
                                    className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                                />
                                <div className="flex items-center space-x-2">
                                    <Monitor className="h-4 w-4 text-slate-600" />
                                    <span className="text-sm font-medium text-slate-700">Console Logging</span>
                                </div>
                            </label>
                            <p className="mt-1 text-xs text-slate-600 ml-6">
                                Output logs to console/terminal in addition to file.
                            </p>
                        </div>

                        {/* Syslog Output */}
                        <div className="border border-slate-200 rounded-lg p-3">
                            <label className="flex items-center space-x-3">
                                <input
                                    type="checkbox"
                                    {...register('logging.enable_syslog')}
                                    className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                                />
                                <div className="flex items-center space-x-2">
                                    <FileText className="h-4 w-4 text-slate-600" />
                                    <span className="text-sm font-medium text-slate-700">System Log (Syslog)</span>
                                </div>
                            </label>
                            <p className="mt-1 text-xs text-slate-600 ml-6">
                                Send logs to system log service for centralized monitoring.
                            </p>
                        </div>
                    </div>
                </div>

                {/* Security Audit Logging */}
                <div className="border border-amber-200 rounded-lg p-4 bg-amber-50">
                    <div className="flex items-center space-x-2 mb-3">
                        <Shield className="h-4 w-4 text-amber-600" />
                        <h4 className="text-sm font-medium text-amber-900">Security Audit Logging</h4>
                    </div>

                    <div className="space-y-4">
                        <label className="flex items-center space-x-3">
                            <input
                                type="checkbox"
                                {...register('logging.enable_audit')}
                                className="rounded border-amber-300 text-amber-600 focus:ring-amber-500"
                            />
                            <span className="text-sm font-medium text-amber-900">Enable Security Audit Logging</span>
                        </label>
                        <p className="text-xs text-amber-800 ml-6">
                            Track security events, connection attempts, and authentication activities for compliance.
                        </p>

                        {enableAudit && (
                            <div className="mt-3">
                                <label className="block text-sm font-medium text-amber-900 mb-2">
                                    Audit Log File Path
                                </label>
                                <input
                                    type="text"
                                    {...register('logging.audit_file')}
                                    placeholder="logs/security_audit.log"
                                    className="w-full px-3 py-2 border border-amber-300 rounded-lg focus:ring-2 focus:ring-amber-500 focus:border-amber-500 text-amber-900 bg-amber-50"
                                />
                                <p className="mt-1 text-xs text-amber-700">
                                    Separate file for security events. Required for compliance in regulated environments.
                                </p>
                            </div>
                        )}
                    </div>
                </div>

                {/* Log Rotation Toggle */}
                <div>
                    <label className="flex items-center space-x-3">
                        <input
                            type="checkbox"
                            {...register('logging.rotation_enabled')}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm font-medium text-slate-700">Enable Log Rotation</span>
                    </label>
                    <p className="mt-1 text-xs text-slate-600 ml-6">
                        Automatically rotate log files when they reach maximum size to prevent disk space issues.
                    </p>
                </div>

                {/* Logging Best Practices */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <h4 className="text-sm font-medium text-blue-900 mb-2">Logging Best Practices</h4>
                    <ul className="text-sm text-blue-800 space-y-1">
                        <li>• Use <strong>info</strong> level for production, <strong>debug</strong> for development</li>
                        <li>• Enable audit logging for healthcare and financial deployments</li>
                        <li>• Set up log rotation to prevent disk space issues</li>
                        <li>• Monitor log files for security events and performance issues</li>
                    </ul>
                </div>
            </div>
        </div>
    );
} 