const fs = require('fs');
const path = require('path');

// Get the docs/website directory (parent of scripts)
// __dirname will be scripts/ when running from package.json
const docsWebsiteDir = path.resolve(__dirname, '..');
const versionsFile = path.resolve(docsWebsiteDir, 'versions.json');
const translationsSource = path.resolve(docsWebsiteDir, 'i18n', 'it', 'docusaurus-plugin-content-docs', 'current');
const currentJsonSource = path.resolve(docsWebsiteDir, 'i18n', 'it', 'docusaurus-plugin-content-docs', 'current.json');
const i18nBase = path.resolve(docsWebsiteDir, 'i18n', 'it', 'docusaurus-plugin-content-docs');

// Read versions
let versions = [];
if (fs.existsSync(versionsFile)) {
  try {
    versions = JSON.parse(fs.readFileSync(versionsFile, 'utf8'));
  } catch (error) {
    console.error('Error reading versions.json:', error);
    process.exit(1);
  }
}

// Also check for versioned_docs directories if versions.json is empty or incomplete
if (versions.length === 0) {
  const versionedDocsPath = path.resolve(docsWebsiteDir, 'versioned_docs');
  if (fs.existsSync(versionedDocsPath)) {
    try {
      const entries = fs.readdirSync(versionedDocsPath, { withFileTypes: true });
      versions = entries
        .filter(entry => entry.isDirectory() && entry.name.startsWith('version-'))
        .map(entry => entry.name.replace('version-', ''));
      console.log(`Found ${versions.length} versions from versioned_docs directory: ${versions.join(', ')}`);
    } catch (error) {
      console.warn(`Warning: Could not read versioned_docs directory: ${error.message}`);
    }
  }
}

if (versions.length === 0) {
  console.log('No versions found. Skipping translation sync.');
  process.exit(0);
}

// Get all files from current translations (recursively for subdirectories)
function getAllFiles(dir, baseDir = dir) {
  const files = [];
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    const relativePath = path.relative(baseDir, fullPath);
    
    if (entry.isDirectory()) {
      // Recursively get files from subdirectories
      files.push(...getAllFiles(fullPath, baseDir));
    } else {
      files.push(relativePath);
    }
  }
  
  return files;
}

const sourceFiles = getAllFiles(translationsSource);

console.log(`Syncing translations from 'current' to ${versions.length} versions...`);

// Copy translations to each version
versions.forEach(version => {
  const versionDir = path.join(i18nBase, `version-${version}`);
  
  // Create version directory if it doesn't exist
  if (!fs.existsSync(versionDir)) {
    fs.mkdirSync(versionDir, { recursive: true });
    console.log(`Created directory: ${versionDir}`);
  }
  
  // Copy each file from current directory
  sourceFiles.forEach(file => {
    const sourceFile = path.join(translationsSource, file);
    const destFile = path.join(versionDir, file);
    const destDir = path.dirname(destFile);
    
    // Create subdirectory if needed
    if (!fs.existsSync(destDir)) {
      fs.mkdirSync(destDir, { recursive: true });
    }
    
    // Only copy if the file does not already exist in this version directory.
    // This preserves version-specific content and prevents current docs from
    // overwriting old versions with links to pages that didn't exist yet.
    if (!fs.existsSync(destFile)) {
      fs.copyFileSync(sourceFile, destFile);
    }
  });
  console.log(`  version-${version}: synced (new files only, existing files preserved)`);
  
  // Copy and update current.json if it exists (it's in the parent directory, not in current/)
  // Docusaurus expects the JSON file at the same level as the version directory, not inside it
  if (fs.existsSync(currentJsonSource)) {
    const destFileName = `version-${version}.json`;
    const finalDestFile = path.join(i18nBase, destFileName);
    
    // Read and update the JSON content
    const content = JSON.parse(fs.readFileSync(currentJsonSource, 'utf8'));
    // Update version.label message if needed
    if (content['version.label']) {
      content['version.label'].message = version;
    }
    
    fs.writeFileSync(finalDestFile, JSON.stringify(content, null, 2) + '\n');
    console.log(`  version-${version}: ${destFileName} updated`);
  }
});

console.log('Translation sync completed!');
