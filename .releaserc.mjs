// @ts-check

/** @type {import('semantic-release').GlobalConfig} */
export default {
    preset: 'angular',
    tagFormat: 'v${version}',
    repositoryUrl: 'https://github.com/telekom/controlplane-mono.git',
    branches: [
        'master',
        'main',
        'next',
        'next-major',
    ],
    plugins: [
        '@semantic-release/commit-analyzer',
        '@semantic-release/release-notes-generator',
        '@semantic-release/changelog',
        ['@semantic-release/git', {
            assets: [
                'CHANGELOG.md',
                'deploy/kustomization.yaml',
            ],
        }],
        ['@semantic-release/exec', {
            prepareCmd: `bash ./update_install.sh "\${nextRelease.gitTag}"`,
            publishCmd: `echo "\${nextRelease.notes}" > /tmp/release-notes.md`,
        }],
    ],
};
