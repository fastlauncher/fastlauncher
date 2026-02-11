{
  buildGoModule
}:
buildGoModule {
  name = "fastlauncher";
  src = ./.;
  vendorHash = "sha256-2z76m6w4gClmboHeBXxMO8xq+wqXo78O1U2yxuP1sMk=";
  meta.mainProgram = "fastlauncher";
}
